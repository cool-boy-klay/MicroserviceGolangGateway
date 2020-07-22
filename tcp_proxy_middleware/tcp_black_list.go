package tcp_proxy_middleware

import (
	"fmt"
	"github.com/cool-boy-klay/go_gateway/dao"
	"github.com/cool-boy-klay/go_gateway/public"
	"strings"
)

//http白名单
func TCPBlackListMiddleware() func(c *TcpSliceRouterContext) {
	return func(c *TcpSliceRouterContext) {
		serverInterface:=c.Get("service")
		if serverInterface==nil{
			c.conn.Write([]byte("get service empty"))
			c.Abort()

			return
		}
		serviceDetail:=serverInterface.(*dao.ServiceDetail)

		whiteIplist:=[]string{}
		if serviceDetail.AccessControl.WhiteList!=""{
			whiteIplist=strings.Split(serviceDetail.AccessControl.WhiteList,",")
		}
		blackIplist:=[]string{}
		if serviceDetail.AccessControl.BlackList!=""{
			blackIplist=strings.Split(serviceDetail.AccessControl.BlackList,",")
		}

		splits:=strings.Split(c.conn.RemoteAddr().String(),":")
		clientIP:=""
		if len(splits)==2{
			clientIP = splits[0]
		}
		if serviceDetail.AccessControl.OpenAuth==1&&len(whiteIplist)==0&&len(blackIplist)>0{
			if public.InStringSlice(blackIplist,clientIP){
				c.conn.Write([]byte(fmt.Sprintf("%s  in black ip list",clientIP)))
				c.Abort()
				return
			}


		}
		c.Next()

	}
}

