package tcp_proxy_middleware

import (
	"fmt"
	"github.com/cool-boy-klay/go_gateway/dao"
	"github.com/cool-boy-klay/go_gateway/public"
	"strings"
)

//http白名单
func TCPWhiteListMiddleware() func(c *TcpSliceRouterContext) {
	return func(c *TcpSliceRouterContext) {
		serverInterface:=c.Get("service")
		if serverInterface==nil{
			c.conn.Write([]byte("service not found"))
			c.Abort()
			return
		}
		serviceDetail:=serverInterface.(*dao.ServiceDetail)

		iplist:=[]string{}
		if serviceDetail.AccessControl.WhiteList!=""{
			iplist=strings.Split(serviceDetail.AccessControl.WhiteList,",")
		}

		splits:=strings.Split(c.conn.RemoteAddr().String(),":")
		clientIP:=""
		if len(splits)==2{
			clientIP = splits[0]
		}
		if serviceDetail.AccessControl.OpenAuth==1&&len(iplist)>0{
			if !public.InStringSlice(iplist,clientIP){
				c.conn.Write([]byte(fmt.Sprintf("%s not int white ip list",clientIP)))
			}
			c.Abort()
			return

		}
		c.Next()

	}
}

