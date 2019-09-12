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

	onConnected  func(c *Client)
	onConnecting func(c *Client)
	onTimeout    func(c *Client)
	onError      func(c *Client, err error)

	onRelayOpen   func(data []byte)
	onRelayClosed func(data []byte)
	onRelayReset  func(data []byte)
}

func NewSDK(addr string) *Client {
	client := &Client{
		addr: addr,
	}
	client.init()
	return client
}

func (c *Client) OnConnecting(f func(c *Client)) {
	c.onConnecting = f
}

func (c *Client) OnConnected(f func(c *Client)) {
	c.onConnected = f
}

func (c *Client) OnTimeout(f func(c *Client)) {
	c.onTimeout = f
}

func (c *Client) OnError(f func(c *Client, err error)) {
	c.onError = f
}

func (c *Client) OnRelayOpen(f func(data []byte)) {
	c.onRelayOpen = f
}

func (c *Client) OnRelayClosed(f func(data []byte)) {
	c.onRelayClosed = f
}

func (c *Client) OnRelayReset(f func(data []byte)) {
	c.onRelayReset = f
}

func (c *Client) init() {

	c.heartBeatTicker = time.NewTicker(time.Second * 5)
	c.heartBeatTickerOver = make(chan interface{})

	c.reconnectTicker = time.NewTicker(time.Second * 15)
	c.reconnectTickerOver = make(chan interface{})

	go c.connect()
	go c.heartBeat()
}

func (c *Client) connect() {
	if !c.connected {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:17000", c.addr), time.Second*3)
		if c.onConnecting != nil {
			c.onConnecting(c)
		}
		if err != nil {
			if c.onTimeout != nil {
				c.onTimeout(c)
			}
			log.Fatal("连接服务器失败", err)
			return
		}
		tcpConn, ok := conn.(*net.TCPConn)
		if !ok {
			if c.onError != nil {
				c.onError(c, err)
			}
			log.Fatal("连接类型转换失败", err)
			return
		}
		c.sess = tcpConn
		err = tcpConn.SetWriteBuffer(5000)
		err2 := tcpConn.SetReadBuffer(5000)
		if err != nil || err2 != nil {
			if c.onError != nil {
				c.onError(c, err)
			}
			c.connected = false
			log.Fatal("设置读写缓冲区失败", err)
		}
		if c.onConnected != nil {
			c.connected = true
			c.onConnected(c)
		}
		go c.clientHandle()
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
				log.Println("客户端发送心跳包", data[4])
				if err != nil {
					if c.onError != nil {
						c.onError(c, err)
					}
					log.Println("发送心跳包失败")
					return
				}
			}
		case <-c.heartBeatTickerOver:
			log.Println("停止心跳")
		}
	}
}

func (c *Client) clientHandle() {
	defer func() {
		if c.sess != nil {
			c.Handle(DisconnectID, nil, c.sess)
			err := c.sess.Close()
			if err != nil {
				if c.onError != nil {
					c.onError(c, err)
				}
				log.Println("关闭连接失败")
			}
		}
	}()

	buf := make([]byte, 1024)
	var cache bytes.Buffer
	for {
		n, err := c.sess.Read(buf)
		if err != nil {
			break
		}
		cache.Write(buf[:n])
		for {
			if c.unpack(&cache, c.sess) {
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
			if c.onError != nil {
				c.onError(c, err)
			}
			log.Println("关闭连接失败", err)
		}
	}
	c.connected = false
	c.sess = nil
	log.Println("客户端关闭连接成功")
}

func (c *Client) RelayOpen(relay []bool) {
	if c.connected {
		log.Println("RelayOpen")
		if len(relay) != 16 {
			log.Println("打开继电器参数长度必须为16")
			return
		}
		open := OpenMessageRequest{
			Relay: relay,
		}
		err := c.Send(&open)
		if err != nil {
			if c.onError != nil {
				c.onError(c, err)
			}
		}
	} else {
		log.Println("请重新连接服务器")
	}
}

func (c *Client) RelayClosed(relay []bool) {
	if c.connected {
		if len(relay) != 16 {
			log.Println("打开继电器参数长度必须为16")
			return
		}
		open := CloseMessageRequest{
			Relay: relay,
		}
		encode, _ := Encode(&open)
		_, err := c.sess.Write(encode)
		if err != nil {
			if c.onError != nil {
				c.onError(c, err)
			}
		}
	} else {
		log.Println("请重新连接服务器")
	}

}

func (c *Client) RelayReset() {
	if c.connected {
		open := ResetMessageRequest{}
		encode, _ := Encode(&open)
		_, err := c.sess.Write(encode)
		if err != nil {
			if c.onError != nil {
				c.onError(c, err)
			}
		}
	} else {
		log.Println("请重新连接服务器")
	}
}
