# relaySDK
> 继电器客户端SDK

## Example

```go
package main

import (
	"github.com/Atian-OE/RelaySDK_Golang"
    "log"
    "sync"
    "testing"
    "time"
)
func main()  {
	   c := relaysdk.NewSDKClient("192.168.0.112").
       		SetReconnectTime(5).
       		SetReconnectTimes(5).
       		OnTimeout(func(c *relaysdk.Client) {
       			log.Println("连接超时........")
       			c.SetAddress("192.168.0.113").OnConnected(func(c *relaysdk.Client) {
       				c.RelayCloseAll()
       				c.RelayOpenAll()
       			}).OnTimeout(func(c *relaysdk.Client) {
       				c.SetAddress("192.168.0.111").OnConnected(func(c *relaysdk.Client) {
       					log.Println("连接成功......开始逻辑处理......")
       					c.RelayCloseAll()
       					time.Sleep(time.Second*3)
       					c.RelayOpenAll()
       					time.Sleep(time.Second*3)
       					c.RelayCloseAll()
       					time.Sleep(time.Second*5)
       					c.Close()
       				})
       			})
       		}).OnConnected(func(c *relaysdk.Client) {
       	})
       	log.Println(c.Id())
       	time.Sleep(time.Minute/2)
       	log.Println("单条测试关闭.....")
}

```