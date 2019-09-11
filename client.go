package relaySDK

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net"
	"time"
)

type Client struct {
	sess      *net.TCPConn
	connected bool
	addr      string

	reconnectTicker     *time.Ticker     //自动连接
	reconnectTickerOver chan interface{} //关闭自动连接

	heartBeatTicker     *time.Ticker     //心跳包的发送
	heartBeatTickerOver chan interface{} //关闭心跳
}

func NewSDK(addr string) *Client {
	client := &Client{
		addr: addr,
	}
	client.init()
	return client
}

func (c *Client) init() {

	c.heartBeatTicker = time.NewTicker(time.Second * 5)
	c.heartBeatTickerOver = make(chan interface{})

	c.reconnectTicker = time.NewTicker(time.Second * 15)
	c.reconnectTickerOver = make(chan interface{})

	c.connect()
}

func (c *Client) connect() {
	if !c.connected {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:17000", c.addr), time.Second*3)
		if err != nil {
			log.Fatal("连接服务器失败", err)
			return
		}
		tcpConn, ok := conn.(*net.TCPConn)
		if !ok {
			log.Fatal("连接类型转换失败", err)
			return
		}
		c.sess = tcpConn
		err = tcpConn.SetWriteBuffer(5000)
		err2 := tcpConn.SetReadBuffer(5000)
		if err != nil || err2 != nil {
			log.Fatal("设置读写缓冲区失败", err)
		}
		go c.clientHandle(tcpConn)
	}
}

func (c *Client) reconnect() {
	c.connected = false
	c.connect()
	for {
		select {
		case <-c.reconnectTicker.C:
			c.connect()

		case <-c.reconnectTickerOver:
			return
		}
	}
}

func (c *Client) heartBeat() {
	for {
		select {
		case <-c.heartBeatTicker.C:
			if c.connected {
				data, _ := Encode(&HeartBeat{})
				_, err := c.sess.Write(data)
				if err != nil {
					log.Println("发送心跳包失败")
					return
				}
			}

		case <-c.heartBeatTickerOver:
			return
		}
	}
}

func (c *Client) clientHandle(conn net.Conn) {
	c.Handle(ConnectID, nil, conn)
	defer func() {
		if conn != nil {
			c.Handle(DisconnectID, nil, conn)
			err := conn.Close()
			if err != nil {
				log.Println("关闭连接失败")
			}
		}
	}()

	buf := make([]byte, 1024)
	var cache bytes.Buffer
	for {
		n, err := conn.Read(buf)
		if err != nil {
			break
		}

		cache.Write(buf[:n])
		for {
			if c.unpack(&cache, conn) {
				break
			}
		}
	}
}

func (c *Client) unpack(cache *bytes.Buffer, conn net.Conn) bool {
	if cache.Len() < 5 {
		return true
	}
	buf := cache.Bytes()
	pkgSize := ByteToInt(buf[:4])
	if pkgSize > len(buf)-5 {
		return true
	}
	cmd := buf[4]
	c.Handle(MsgID(cmd), buf[:pkgSize+5], conn)
	cache.Reset()
	cache.Write(buf[5+pkgSize:])
	return false
}

func (c *Client) Send(msg interface{}) error {
	b, err := Encode(msg)
	if err != nil {
		return err
	}
	if !c.connected {
		return errors.New("client not connected")
	}
	_, err = c.sess.Write(b)
	return err
}

func (c *Client) Close() {
	c.reconnectTicker.Stop()
	c.reconnectTickerOver <- false
	close(c.reconnectTickerOver)

	c.heartBeatTicker.Stop()
	c.heartBeatTickerOver <- false
	close(c.heartBeatTickerOver)

	if c.sess != nil {
		err := c.sess.Close()
		if err != nil {
			log.Println("关闭连接失败", err)
		}
	}
	c.sess = nil
}
