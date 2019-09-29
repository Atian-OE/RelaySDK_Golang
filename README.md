# relaySDK
> 继电器客户端SDK

## Example

```go
package main

import (
    "github.com/Atian-OE/RelaySDK_Golang/relaySDK"
    "log"
    "sync"
    "testing"
    "time"
)
func main()  {
	    sdkClient := NewSDKClient("192.168.0.176:17000")
    
    	sdkClient.OnConnecting(func(c *Client) {
    		log.Println("正在连接到服务器")
    	})
    
    	sdkClient.OnConnected(func(c *Client) {
    		log.Println("已连接到服务器")
    	})
    
    	sdkClient.OnTimeout(func(c *Client) {
    		log.Println("连接到服务器超时")
    		return
    	})
    
    	sdkClient.OnError(func(c *Client, err error) {
    		log.Println("err", err)
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

```