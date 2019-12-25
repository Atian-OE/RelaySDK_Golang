package relaysdk

import (
	"fmt"
	"log"
	"net"
)

func (c *Client) Handle(msgId MsgID, data []byte, conn net.Conn) {
	switch msgId {
	case ConnectID:
		log.Println(fmt.Sprintf("[ 继电器客户端%s ]连接成功...", c.Id()))
	case DisconnectID:
		log.Println(fmt.Sprintf("[ 继电器客户端%s ]正在关闭连接...", c.Id()))
		if c.onDisconnect != nil {
			go c.onDisconnect(c)
		}
	case Open:
		if c.onRelayOpen != nil {
			c.onRelayOpen(data[5:])
		}
	case Close:
		if c.onRelayClosed != nil {
			c.onRelayClosed(data[5:])
		}
	case Reset:
		if c.onRelayReset != nil {
			c.onRelayReset(data[5:])
		}
	case HeartBeatID:
		log.Println(fmt.Sprintf("[ 继电器客户端%s ]正在心跳💓...", c.Id()))
	default:
		log.Println(string(data))
	}
}
