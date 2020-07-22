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
func HttpWhiteListMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		serverInterface,ok:=c.Get("service")
		if !ok{
			middleware.ResponseError(c,2001,errors.New("service not found"))
			c.Abort()
			return
		}
		serviceDetail:=serverInterface.(*dao.ServiceDetail)

		iplist:=[]string{}
		if serviceDetail.AccessControl.WhiteList!=""{
			iplist=strings.Split(serviceDetail.AccessControl.WhiteList,",")
		}

		if serviceDetail.AccessControl.OpenAuth==1&&len(iplist)>0{
			if !public.InStringSlice(iplist,c.ClientIP()){
				middleware.ResponseError(c,3001,errors.New(fmt.Sprintf("%s not int white ip list",c.ClientIP())))
			}
			c.Abort()
			return

		}
		c.Next()

	}
}

