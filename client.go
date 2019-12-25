package relaysdk

import (
	"bytes"
	"errors"
	"fmt"
	uuid "github.com/iris-contrib/go.uuid"
	"log"
	"net"
	"time"
)

//实例化客户端
func NewSDKClient(addr string) *Client {
	c := &Client{
		addr: addr,
		port: 17000,
	}
	if c.Id() == "" {
		v4, err := uuid.NewV4()
		if err != nil {
			c.SetId("")
		} else {
			c.SetId(v4.String()[:8])
		}
	}
	c.init()
	return c
}

//初始化
func (c *Client) init() {
	c.reconnectTicker = time.NewTicker(time.Second * c.ReconnectTime())
	c.reconnectTickerOver = make(chan interface{})
	go c.reconnect()
}

func (c *Client) reconnect() {
	c.connected = false
	c.connect()
	for {
		select {
		case <-c.reconnectTicker.C:
			go func() {
				if !c.IsConnected() {
					c.count += 1
					if c.ReconnectTimes() == 0 {
						log.Println(fmt.Sprintf("[ 继电器客户端%s ]正在无限尝试第[ %d/%d ]次重新连接[ %s ]...", c.Id(), c.count, c.ReconnectTimes(), c.addr))
						c.connect()
					} else {
						if c.count <= c.ReconnectTimes() {
							log.Println(fmt.Sprintf("[ 继电器客户端%s ]正在尝试第[ %d/%d ]次连接[ %s ]...", c.Id(), c.count, c.ReconnectTimes(), c.addr))
							c.connect()
						} else {
							log.Println(fmt.Sprintf("[ 继电器客户端%s ]第[ %d/%d ]次连接失败,断开连接[ %s ]...", c.Id(), c.count-1, c.ReconnectTimes(), c.addr))
							c.Close()
						}
					}
				}
			}()

		case <-c.reconnectTickerOver:
			log.Println(fmt.Sprintf("[ 继电器客户端%s ]断开连接[ %s ]...", c.Id(), c.addr))
			return
		}
	}
}

//连接
func (c *Client) connect() {
	if !c.connected {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", c.addr, c.Port()), time.Second*3)
		if c.onConnecting != nil {
			go c.onConnecting(c)
		}
		if err != nil {
			if c.onTimeout != nil {
				go c.onTimeout(c)
			}
			return
		}
		tcpConn, ok := conn.(*net.TCPConn)
		if !ok {
			return
		}
		c.sess = tcpConn
		err = tcpConn.SetWriteBuffer(5000)
		err2 := tcpConn.SetReadBuffer(5000)
		if err != nil || err2 != nil {
			return
		}
		c.connected = true
		if c.onConnected != nil {
			go c.onConnected(c)
		}
		go c.heartBeat()
		go c.clientHandle()
	}
}

//心跳
func (c *Client) heartBeat() {
	c.heartBeatTicker = time.NewTicker(time.Second * c.HeartBeatTime())
	c.heartBeatTickerOver = make(chan interface{}, 0)
	for {
		select {
		case <-c.heartBeatTicker.C:
			if c.connected {
				data, _ := Encode(&HeartBeat{})
				_, err := c.sess.Write(data)
				if err != nil {
					log.Println(fmt.Sprintf("[ 继电器客户端%s ]发送心跳包失败...", c.Id()))
				}
			}
		case <-c.heartBeatTickerOver:
			log.Println(fmt.Sprintf("[ 继电器客户端%s ]正在停止心跳💓...", c.Id()))
			return
		}
	}
}

//消息处理,解包
func (c *Client) clientHandle() {
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

//解包
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

//发送数据
func (c *Client) Send(msg interface{}) error {
	b, err := Encode(msg)
	if err != nil {
		return err
	}
	if c.IsConnected() {
		_, err = c.sess.Write(b)
		return err
	}
	return errors.New(fmt.Sprintf("[ 继电器客户端%s ]未连接[ %s ]...", c.Id(), c.addr))
}

//关闭操作
func (c *Client) Close() {
	if c.sess != nil {
		c.Handle(DisconnectID, nil, c.sess)
		c.heartBeatTickerOver <- false
		c.reconnectTickerOver <- false
		c.heartBeatTicker.Stop()
		close(c.heartBeatTickerOver)
		c.reconnectTicker.Stop()
		close(c.reconnectTickerOver)
		err := c.sess.Close()
		if err != nil {
			log.Println(fmt.Sprintf("[ 继电器客户端%s ]关闭失败:[ %s ]...", c.Id(), err))
		}
		c.sess = nil
	}
	c.connected = false
	log.Println(fmt.Sprintf("[ 继电器客户端%s ]关闭成功...", c.Id()))
}

//闭合所有继电器
func (c *Client) RelayOpenAll() {
	_ = c.Send(&OpenMessageRequest{
		Relay: []bool{
			true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true,
		},
	})
}

//打开继电器
func (c *Client) RelayOpen(relay []bool) {
	if len(relay) != 32 {
		log.Println("打开继电器参数长度必须为32")
		return
	}
	_ = c.Send(&OpenMessageRequest{
		Relay: relay,
	})
}

//断开所有继电器
func (c *Client) RelayCloseAll() {
	_ = c.Send(&CloseMessageRequest{
		Relay: []bool{
			true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true,
			true, true, true, true, true, true, true, true,
		},
	})
}

//关闭继电器
func (c *Client) RelayClosed(relay []bool) {
	if len(relay) != 32 {
		log.Println("打开继电器参数长度必须为32")
		return
	}
	encode, _ := Encode(&CloseMessageRequest{
		Relay: relay,
	})
	_ = c.Send(encode)
}

//重置继电器
func (c *Client) RelayReset() {
	encode, _ := Encode(&ResetMessageRequest{})
	_ = c.Send(encode)
}
