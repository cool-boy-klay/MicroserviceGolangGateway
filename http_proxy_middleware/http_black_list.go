package http_proxy_middleware

import (
	"fmt"
	"github.com/cool-boy-klay/go_gateway/dao"
	"github.com/cool-boy-klay/go_gateway/middleware"
	"github.com/cool-boy-klay/go_gateway/public"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"strings"
)

//http白名单
func HttpBlackListMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		serverInterface,ok:=c.Get("service")
		if !ok{
			middleware.ResponseError(c,2001,errors.New("service not found"))
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

		if serviceDetail.AccessControl.OpenAuth==1&&len(whiteIplist)==0&&len(blackIplist)>0{
			if public.InStringSlice(blackIplist,c.ClientIP()){
				middleware.ResponseError(c,3001,errors.New(fmt.Sprintf("%s  in black ip list",c.ClientIP())))
				c.Abort()
				return
			}


		}
		c.Next()

	}
}

