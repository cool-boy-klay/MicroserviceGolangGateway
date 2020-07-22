package tcp_proxy_middleware

import (
	"fmt"
	"github.com/cool-boy-klay/go_gateway/dao"
	"github.com/cool-boy-klay/go_gateway/public"
	"time"
)

//流量统计
func TCPFlowCountMiddleware() func(c *TcpSliceRouterContext) {
	return func(c *TcpSliceRouterContext) {
		serverInterface:=c.Get("service")
		if serverInterface==nil{
			c.conn.Write([]byte("service not found"))
			c.Abort()
			return
		}
		serviceDetail:=serverInterface.(*dao.ServiceDetail)
		//统计全站
		totalCounter,err:=public.FlowCounterHandler.GetFlowCounter(public.FlowTotal)
		if err!=nil{
			c.conn.Write([]byte(err.Error()))
			c.Abort()
			return
		}
		totalCounter.Increase()

		dayCount,_:=totalCounter.GetDayData(time.Now())
		fmt.Printf("totalCount qps:%s,dayCount:%v",totalCounter.QPS,dayCount)

		//统计服务
		ServiceCounter,err:=public.FlowCounterHandler.GetFlowCounter(public.FlowServicePrefix+serviceDetail.Info.ServiceName)
		if err!=nil{
			c.conn.Write([]byte(err.Error()))
			c.Abort()
			return
		}
		ServiceCounter.Increase()
		dayServiceCount,_:=ServiceCounter.GetDayData(time.Now())
		fmt.Printf("totalCount qps:%s,dayCount:%v",ServiceCounter.QPS,dayServiceCount)


		c.Next()
	}
}
