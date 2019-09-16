package relaySDK

import (
	"log"
	"sync"
	"testing"
	"time"
)

func TestClient(t *testing.T) {
	group := sync.WaitGroup{}
	group.Add(1)
	sdkClient := NewSDKClient("192.168.0.176")

	sdkClient.OnConnecting(func(c *Client) {
		log.Println("正在连接到服务器")
	})

	sdkClient.OnConnected(func(c *Client) {
		log.Println("已连接到服务器")
	})

	sdkClient.OnTimeout(func(c *Client) {
		log.Println("连接到服务器超时")
		group.Done()
		return
	})

	sdkClient.OnError(func(c *Client, err error) {
		log.Println("err", err)
		group.Done()
		return
	})
	time.Sleep(time.Second * 2)
	open := []bool{true, false, true, false, true, false, true, false, true, false, true, false, true, false, true, true}
	sdkClient.RelayOpen(open)
	sdkClient.OnRelayOpen(func(data []byte) {
		log.Println("OnRelayOpen", string(data))
	})
	time.Sleep(3 * time.Second)
	closed := []bool{false, false, true, false, true, false, true, false, true, false, true, false, true, false, true, true}
	sdkClient.RelayClosed(closed)
	sdkClient.OnRelayClosed(func(data []byte) {
		log.Println("OnRelayClosed", string(data))
	})
	time.Sleep(3 * time.Second)
	sdkClient.RelayReset()
	sdkClient.OnRelayReset(func(data []byte) {
		log.Println("OnRelayReset", string(data))
	})
	time.Sleep(3 * time.Second)
	group.Done()
	go sdkClient.Close()
	group.Wait()
}
