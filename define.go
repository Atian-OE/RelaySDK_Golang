package relaysdk

import (
	"net"
	"time"
)

//客户端
type Client struct {
	sess         *net.TCPConn
	id           string
	connected    bool
	count        int
	reconnecting bool //正在重新连接的标志
	addr         string
	port         int //端口 默认17083

	reconnectTime       time.Duration //重连时间
	reconnectTimes      int
	reconnectTicker     *time.Ticker     //自动连接
	reconnectTickerOver chan interface{} //关闭自动连接

	heartBeatTime       time.Duration    //重连时间
	heartBeatTicker     *time.Ticker     //心跳包的发送
	heartBeatTickerOver chan interface{} //关闭心跳

	onDisconnect func(c *Client)
	onConnected  func(c *Client)
	onConnecting func(c *Client)
	onTimeout    func(c *Client)

	onRelayOpen   func(data []byte)
	onRelayClosed func(data []byte)
	onRelayReset  func(data []byte)
}

func (c *Client) IsReconnecting() bool {
	return c.reconnecting
}

func (c *Client) IsConnected() bool {
	return c.connected
}

func (c *Client) Id() string {
	return c.id
}

func (c *Client) SetId(id string) *Client {
	c.id = id
	return c
}

func (c *Client) Port() int {
	return c.port
}

func (c *Client) SetPort(port int) *Client {
	c.port = port
	return c
}

func (c *Client) Address() string {
	return c.addr
}

func (c *Client) SetAddress(address string) *Client {
	c.addr = address
	return c
}

func (c *Client) ReconnectTimes() int {
	return c.reconnectTimes
}

func (c *Client) SetReconnectTimes(reconnectTimes int) *Client {
	c.reconnectTimes = reconnectTimes
	return c
}

func (c *Client) ReconnectTime() time.Duration {
	if c.reconnectTime == 0 {
		c.reconnectTime = 10
	}
	return c.reconnectTime
}

func (c *Client) SetReconnectTime(reconnectTime time.Duration) *Client {
	c.reconnectTime = reconnectTime
	return c
}

func (c *Client) HeartBeatTime() time.Duration {
	if c.heartBeatTime == 0 {
		c.heartBeatTime = 5
	}
	return c.heartBeatTime
}

func (c *Client) OnConnecting(f func(c *Client)) *Client {
	c.onConnecting = f
	return c
}

func (c *Client) OnConnected(f func(c *Client)) *Client {
	c.onConnected = f
	return c
}

func (c *Client) OnDisconnect(f func(c *Client)) *Client {
	c.onDisconnect = f
	return c
}

func (c *Client) OnTimeout(f func(c *Client)) *Client {
	c.onTimeout = f
	return c
}

func (c *Client) OnRelayOpen(f func(data []byte)) *Client {
	c.onRelayOpen = f
	return c
}

func (c *Client) OnRelayClosed(f func(data []byte)) *Client {
	c.onRelayClosed = f
	return c
}

func (c *Client) OnRelayReset(f func(data []byte)) *Client {
	c.onRelayReset = f
	return c
}
