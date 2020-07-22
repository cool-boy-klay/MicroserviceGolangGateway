package http_proxy_middleware

import (
	"fmt"
	"github.com/cool-boy-klay/go_gateway/dao"
	"github.com/cool-boy-klay/go_gateway/middleware"
	"github.com/cool-boy-klay/go_gateway/public"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)


func HTTPJwtFlowLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		appInterface, ok := c.Get("app")
		if !ok {
			c.Next()
			return
		}
		appInfo := appInterface.(*dao.App)
		if appInfo.Qps > 0 {
			clientLimiter, err := public.FlowLimitHandler.GetFlowLimit(
				public.FlowAppPrefix+appInfo.AppID+"_"+c.ClientIP(),
				float64(appInfo.Qps))
			if err != nil {
				middleware.ResponseError(c, 5001, err)
				c.Abort()
				return
			}
			if !clientLimiter.Allow() {
				middleware.ResponseError(c, 5002, errors.New(fmt.Sprintf("%v flow limit %v", c.ClientIP(), appInfo.Qps), ))
				c.Abort()
				return
			}
		}
		c.Next()
	}
}
