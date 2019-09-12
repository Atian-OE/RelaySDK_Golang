package relaySDK

import (
	"log"
	"sync"
	"testing"
	"time"
)

func TestName3(t *testing.T) {
	log.Printf("%d", byte(250))
	log.Printf("%s", string(byte(250)))
}
func TestName2(t *testing.T) {
	log.Println(time.Now().UnixNano() / 1000000)    //1568269794412
	log.Println(time.Now().UnixNano() / 1000000000) //1568269794412
	log.Println(time.Now().Unix())                  //1568269837
	log.Println(time.Now().UnixNano())              //1568269794412
}

func TestName(t *testing.T) {
	group := sync.WaitGroup{}
	group.Add(1)
	sdk := NewSDK("192.168.0.176")

	sdk.OnConnecting(func(c *Client) {
		log.Println("正在连接到服务器")
	})

	sdk.OnConnected(func(c *Client) {
		log.Println("已连接到服务器")
	})

	sdk.OnTimeout(func(c *Client) {
		log.Println("连接到服务器超时")
		group.Done()
		return
	})

	sdk.OnError(func(c *Client, err error) {
		log.Println("err", err)
		group.Done()
		return
	})
	time.Sleep(time.Second * 2)
	open := []bool{true, false, true, false, true, false, true, false, true, false, true, false, true, false, true, true}
	closed := []bool{false, false, true, false, true, false, true, false, true, false, true, false, true, false, true, true}
	sdk.RelayOpen(open)
	sdk.OnRelayOpen(func(data []byte) {
		log.Println(string(data))
	})
	time.Sleep(3 * time.Second)
	sdk.RelayClosed(closed)
	sdk.OnRelayClosed(func(data []byte) {
		log.Println(string(data))
	})
	time.Sleep(3 * time.Second)
	sdk.RelayReset()
	sdk.OnRelayReset(func(data []byte) {
		log.Println(data)
	})
	time.Sleep(3 * time.Second)
	sdk.Close()
	group.Wait()
}
