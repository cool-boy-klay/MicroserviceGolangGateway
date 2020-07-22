package http_proxy_middleware

import (
	"fmt"
	"github.com/cool-boy-klay/go_gateway/dao"
	"github.com/cool-boy-klay/go_gateway/middleware"
	"github.com/cool-boy-klay/go_gateway/public"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

//租户流量统计
func HttpJwtFlowCountMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		//serverInterface,ok:=c.Get("service")
		//if !ok{
		//	middleware.ResponseError(c,2001,errors.New("service not found"))
		//	c.Abort()
		//	return
		//}
		//serviceDetail:=serverInterface.(*dao.ServiceDetail)

		appInterface,ok:=c.Get("appDetail")
		if !ok{
			c.Next()
			return
		}

		appInfo:=appInterface.(*dao.App)
		appCounter,err:=public.FlowCounterHandler.GetFlowCounter(public.FlowAppPrefix+appInfo.AppID)
		if err!=nil{
			middleware.ResponseError(c,2002,err)
			c.Abort()
			return
		}
		fmt.Println(public.FlowAppPrefix+appInfo.AppID,":",appCounter.TotalCount)
		appCounter.Increase()

		if appInfo.Qpd>0&&appCounter.TotalCount>appInfo.Qpd{
			middleware.ResponseError(c,2003,errors.New(fmt.Sprintf("超出每日请求上限 limit:%v current:%v",appInfo.Qpd,appCounter.TotalCount)))
			c.Abort()
			return
		}
		c.Next()
	}
}
