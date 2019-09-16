# relaySDK
> 继电器客户端SDK

## Example

```go
package main

import (
    "log"
    "relaySDK"
    "sync"
    "testing"
    "time"
)
func main()  {
	sdk := relaySDK.NewSDKClient("192.168.0.176")
 
 	sdk.OnConnecting(func(c *relaySDK.Client) {
 		log.Println("正在连接到服务器")
 	})
 
 	sdk.OnConnected(func(c *relaySDK.Client) {
 		log.Println("已连接到服务器")
 	})
 
 	sdk.OnTimeout(func(c *relaySDK.Client) {
 		log.Println("连接到服务器超时")
 		return
 	})
 
 	sdk.OnError(func(c *relaySDK.Client, err error) {
 		log.Println("err", err)
 		return
 	})

 	time.Sleep(time.Second * 2)
 	open := []bool{true, false, true, false, true, false, true, false, true, false, true, false, true, false, true, true}
 	sdk.RelayOpen(open)
 	sdk.OnRelayOpen(func(data []byte) {
 		log.Println(string(data))
 	})

 	time.Sleep(3 * time.Second)
 	closed := []bool{false, false, true, false, true, false, true, false, true, false, true, false, true, false, true, true}
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
 	go sdk.Close()
}

```