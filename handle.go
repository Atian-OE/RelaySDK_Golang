package relaySDK

import (
	"log"
	"net"
)

func (c *Client) Handle(msgId MsgID, data []byte, conn net.Conn) {
	switch msgId {
	case ConnectID:
		if c.onConnected != nil {
			c.onConnected(c)
		}
		log.Println("ConnectID", string(data[5:]))
	case DisconnectID:
		log.Println("DisconnectID", string(data[5:]))
		c.Close()
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
		log.Println("HeartBeatID", string(data[5:]))
	default:
		log.Println(string(data))
	}
}
