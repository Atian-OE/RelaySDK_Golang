package relaySDK

import (
	"log"
	"net"
)

func (c *Client) Handle(msgId MsgID, data []byte, conn net.Conn) {
	switch msgId {
	case ConnectID:
		log.Println("ConnectID")
	case DisconnectID:
		log.Println("DisconnectID")
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
