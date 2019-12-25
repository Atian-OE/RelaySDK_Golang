package relaysdk

import (
	"bytes"
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
			c.SetId(v4.String())
		}
	}
	c.init()
	return c
}

//初始化
func (c *Client) init() {
	c.heartBeatTicker = time.NewTicker(time.Second * c.HeartBeatTime())
	c.heartBeatTickerOver = make(chan interface{})
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
			if !c.connected {
				c.count += 1
				if c.reconnectTimes == 0 {
					log.Println(fmt.Sprintf("[ 继电器客户端%s ]正在无限尝试第[ %d/%d ]次重新连接[ %s ]...", c.Id(), c.count, c.reconnectTimes, c.addr))
					c.connect()
				} else {
					if c.count <= c.reconnectTimes {
						log.Println(fmt.Sprintf("[ 继电器客户端%s ]正在尝试第[ %d/%d ]次重新连接[ %s ]...", c.Id(), c.count, c.reconnectTimes, c.addr))
						c.connect()
					} else {
						log.Println(fmt.Sprintf("[ 继电器客户端%s ]第[ %d/%d ]次重新连接失败,断开连接[ %s ]...", c.Id(), c.count-1, c.reconnectTimes, c.addr))
						c.Close()
					}
				}
			}
		case <-c.reconnectTickerOver:
			log.Println(fmt.Sprintf("[ 继电器客户端%s ]断开连接[ %s ]...", c.Id(), c.addr))
			return
		}
	}
}

//连接
func (c *Client) connect() {
	if !c.connected {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", c.addr, c.port), time.Second*3)
		if c.onConnecting != nil {
			c.onConnecting(c)
		}
		if err != nil {
			if c.onTimeout != nil {
				c.onTimeout(c)
			}
			return
		}
		tcpConn, ok := conn.(*net.TCPConn)
		if !ok {
			if c.onError != nil {
				c.onError(c, err)
			}
			return
		}
		c.sess = tcpConn
		err = tcpConn.SetWriteBuffer(5000)
		err2 := tcpConn.SetReadBuffer(5000)
		if err != nil || err2 != nil {
			if c.onError != nil {
				c.onError(c, err)
			}
			return
		}
		c.connected = true
		if c.onConnected != nil {
			c.onConnected(c)
		}
		go c.heartBeat()
		go c.clientHandle()
	}
}

//心跳
func (c *Client) heartBeat() {
	for {
		select {
		case <-c.heartBeatTicker.C:
			if c.connected {
				data, _ := Encode(&HeartBeat{})
				_, err := c.sess.Write(data)
				if err != nil {
					if c.onError != nil {
						c.onError(c, err)
					}
					log.Println(fmt.Sprintf("[ 继电器客户端%s ]发送心跳包失败...", c.Id()))
					c.reconnect()
					return
				}
			}
		case <-c.heartBeatTickerOver:
			log.Println(fmt.Sprintf("[ 继电器客户端%s ]停止心跳...", c.Id()))
			return
		}
	}
}

//消息处理,解包
func (c *Client) clientHandle() {
	defer c.Close()
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
	if !c.connected {
		c.reconnect()
		time.Sleep(time.Millisecond * 200)
	}
	_, err = c.sess.Write(b)
	return err
}

//关闭操作
func (c *Client) Close() {
	c.reconnectTicker.Stop()
	c.reconnectTickerOver <- false
	close(c.reconnectTickerOver)

	c.heartBeatTicker.Stop()
	c.heartBeatTickerOver <- false
	close(c.heartBeatTickerOver)

	if c.sess != nil {
		c.Handle(DisconnectID, nil, c.sess)
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
	log.Println(fmt.Sprintf("[ 继电器客户端%s ]关闭成功...", c.Id()))
}

//闭合所有继电器
func (c *Client) RelayOpenAll() {
	if c.connected {
		open := OpenMessageRequest{
			Relay: []bool{
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
			},
		}
		err := c.Send(&open)
		if err != nil {
			if c.onError != nil {
				c.onError(c, SendErr(err))
			}
		}
	} else {
		log.Println("打开继电器失败,请重新连接服务器")
	}
}

//打开继电器
func (c *Client) RelayOpen(relay []bool) {
	if c.connected {
		if len(relay) != 32 {
			log.Println("打开继电器参数长度必须为32")
			return
		}
		open := OpenMessageRequest{
			Relay: relay,
		}
		err := c.Send(&open)
		if err != nil {
			if c.onError != nil {
				c.onError(c, SendErr(err))
			}
		}
	} else {
		log.Println("打开继电器失败,请重新连接服务器")
	}
}

//断开所有继电器
func (c *Client) RelayCloseAll() {
	if c.connected {
		open := CloseMessageRequest{
			Relay: []bool{
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
			},
		}
		encode, _ := Encode(&open)
		_, err := c.sess.Write(encode)
		if err != nil {
			if c.onError != nil {
				c.onError(c, err)
			}
		}
	} else {
		log.Println("关闭继电器失败,请重新连接服务器")
	}
}

//关闭继电器
func (c *Client) RelayClosed(relay []bool) {
	if c.connected {
		if len(relay) != 32 {
			log.Println("打开继电器参数长度必须为32")
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
		log.Println("关闭继电器失败,请重新连接服务器")
	}
}

//重置继电器
func (c *Client) RelayReset() {
	if c.connected {
		reset := ResetMessageRequest{}
		encode, _ := Encode(&reset)
		_, err := c.sess.Write(encode)
		if err != nil {
			if c.onError != nil {
				c.onError(c, err)
			}
		}
	} else {
		log.Println("重置继电器失败,请重新连接服务器")
	}
}
