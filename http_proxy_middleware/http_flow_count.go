package http_proxy_middleware

import (
	"fmt"
	"github.com/cool-boy-klay/go_gateway/dao"
	"github.com/cool-boy-klay/go_gateway/middleware"
	"github.com/cool-boy-klay/go_gateway/public"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"time"
)

//流量统计
func HttpFlowCountMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		serverInterface,ok:=c.Get("service")
		if !ok{
			middleware.ResponseError(c,2001,errors.New("service not found"))
			c.Abort()
			return
		}
		serviceDetail:=serverInterface.(*dao.ServiceDetail)
		//统计全站
		totalCounter,err:=public.FlowCounterHandler.GetFlowCounter(public.FlowTotal)
		if err!=nil{
			middleware.ResponseError(c,4001,err)
			c.Abort()
			return
		}
		totalCounter.Increase()

		dayCount,_:=totalCounter.GetDayData(time.Now())
		fmt.Printf("totalCount qps:%s,dayCount:%v",totalCounter.QPS,dayCount)

		//统计服务
		ServiceCounter,err:=public.FlowCounterHandler.GetFlowCounter(public.FlowServicePrefix+serviceDetail.Info.ServiceName)
		if err!=nil{
			middleware.ResponseError(c,4001,err)
			c.Abort()
			return
		}
		ServiceCounter.Increase()
		dayServiceCount,_:=ServiceCounter.GetDayData(time.Now())
		fmt.Printf("totalCount qps:%s,dayCount:%v",ServiceCounter.QPS,dayServiceCount)


		c.Next()
	}
}
