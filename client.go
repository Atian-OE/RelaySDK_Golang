package relaySDK

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

//客户端
type Client struct {
	one       sync.Once
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

//实例化客户端
func NewSDKClient(addr string) *Client {
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

//初始化
func (c *Client) init() {

	c.heartBeatTicker = time.NewTicker(time.Second * 5)
	c.heartBeatTickerOver = make(chan interface{})

	c.reconnectTicker = time.NewTicker(time.Second * 15)
	c.reconnectTickerOver = make(chan interface{})

	go c.reconnect()
}

//连接
func (c *Client) connect() {
	if !c.connected {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s", c.addr), time.Second*3)
		if c.onConnecting != nil {
			c.onConnecting(c)
		}
		if err != nil {
			if c.onTimeout != nil {
				c.onTimeout(c)
			}
			c.reconnect()
			return
		}
		tcpConn, ok := conn.(*net.TCPConn)
		if !ok {
			if c.onError != nil {
				c.onError(c, err)
			}
			c.reconnect()
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

func (c *Client) reconnect() {
	c.connected = false
	time.Sleep(time.Second)
	c.connect()
	for {
		select {
		case <-c.reconnectTicker.C:
			c.connect()
			return

		case <-c.reconnectTickerOver:
			return
		}
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
				log.Println("客户端发送心跳包", data[4])
				if err != nil {
					if c.onError != nil {
						c.onError(c, err)
					}
					log.Println("发送心跳包失败")
					c.reconnect()
					return
				}
			}
		case <-c.heartBeatTickerOver:
			log.Println("停止心跳")
			return
		}
	}
}

//消息处理,解包
func (c *Client) clientHandle() {
	//defer c.Close()
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
	c.one.Do(func() {
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
		log.Println("客户端关闭连接成功")
	})
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
