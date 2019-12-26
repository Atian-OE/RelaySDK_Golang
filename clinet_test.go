package relaysdk_test

import (
	"github.com/Atian-OE/RelaySDK_Golang"
	"log"
	"math"
	"testing"
	"time"
)

func TestName4(t *testing.T) {
	log.Println(math.MaxUint16)
}
func BenchmarkName(b *testing.B) {
	for i := 0; i < b.N; i++ {
		test()
	}
}
func TestClient(t *testing.T) {
	test()
}
func BenchmarkName2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		test2()
	}
}
func TestName3(t *testing.T) {
	for i := 0; i < 10; i++ {
		test2()
	}
}
func test2() {
	c := relaysdk.NewSDKClient("192.168.0.112").
		SetReconnectTime(time.Second * 5).
		SetReconnectTimes(5).
		OnTimeout(func(c *relaysdk.Client) {
			log.Println("连接超时........")
			c.SetAddress("192.168.0.113").OnConnected(func(c *relaysdk.Client) {
				c.RelayCloseAll()
				c.RelayOpenAll()
			}).OnTimeout(func(c *relaysdk.Client) {
				c.SetAddress("192.168.0.119").OnConnected(func(c *relaysdk.Client) {
					log.Println("连接成功......开始逻辑处理......")
					c.RelayCloseAll()
					time.Sleep(time.Second * 2)
					c.RelayOpenAll()
					time.Sleep(time.Second * 2)
					c.RelayCloseAll()
					time.Sleep(time.Second * 2)
					c.Close()
				})
			})
		}).OnConnected(func(c *relaysdk.Client) {})
	log.Println(c.Id())
	time.Sleep(time.Minute * 2)
	log.Println("单条测试关闭.....")
}
func TestName2(t *testing.T) {
	test2()
}

func TestName(t *testing.T) {
	client := relaysdk.NewSDKClient("127.0.0.1:17000")
	time.Sleep(2 * time.Second)
	client.RelayOpenAll()
	time.Sleep(2 * time.Second)
	client.RelayCloseAll()
}
func test() {
	sdkClient := relaysdk.NewSDKClient("192.168.0.111")

	sdkClient.OnConnecting(func(c *relaysdk.Client) {
		log.Println("正在连接到服务器")
	})

	sdkClient.OnConnected(func(c *relaysdk.Client) {
		log.Println("已连接到服务器")
	})

	sdkClient.OnTimeout(func(c *relaysdk.Client) {
		log.Println("连接到服务器超时")
		return
	})

	time.Sleep(time.Second * 2)
	open := [32]bool{true, false, true, false, true, false, true, false, true, false, true, false, true, false, true, true}
	sdkClient.RelayOpen(open[:])
	sdkClient.OnRelayOpen(func(data []byte) {
		log.Println("OnRelayOpen", string(data))
	})
	time.Sleep(3 * time.Second)
	closed := [32]bool{false, false, true, false, true, false, true, false, true, false, true, false, true, false, true, true}
	sdkClient.RelayClosed(closed[:])
	sdkClient.OnRelayClosed(func(data []byte) {
		log.Println("OnRelayClosed", string(data))
	})
	time.Sleep(3 * time.Second)
	sdkClient.RelayReset()
	sdkClient.OnRelayReset(func(data []byte) {
		log.Println("OnRelayReset", string(data))
	})
	time.Sleep(3 * time.Second)
	sdkClient.Close()
}
