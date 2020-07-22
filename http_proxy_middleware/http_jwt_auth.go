package http_proxy_middleware

import (
	"github.com/cool-boy-klay/go_gateway/dao"
	"github.com/cool-boy-klay/go_gateway/middleware"
	"github.com/cool-boy-klay/go_gateway/public"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"strings"
)

//jwt校验
func HttpJwtAuthTokenMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		serverInterface,ok:=c.Get("service")
		if !ok{
			middleware.ResponseError(c,2001,errors.New("service not found"))
			c.Abort()
			return
		}
		serviceDetail:=serverInterface.(*dao.ServiceDetail)

		matched:=false
		//decode jwt token
		token:=strings.ReplaceAll(c.GetHeader("Authorization"),"Bearer ","")
		if token!=""{
			claims,err:=public.JwtDecode(token)
			if err!=nil{
				middleware.ResponseError(c,2002,errors.New("token now exists"))
				c.Abort()
				return
			}
			appList:=dao.AppManagerHandler.GetAppList()
			for _,appInfo :=range appList{
				if appInfo.AppID==claims.Issuer{
					//appInfo放到gin.context
					c.Set("appDetail",appInfo)
					matched = true
					break
				}
			}

		}
		if serviceDetail.AccessControl.OpenAuth==1&&!matched{
			middleware.ResponseError(c,2002,errors.New("not match valid"))
			c.Abort()
			return
		}

		c.Next()

	}
}

