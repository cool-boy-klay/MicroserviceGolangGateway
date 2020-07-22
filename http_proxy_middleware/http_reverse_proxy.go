package http_proxy_middleware

import (
	"errors"
	"github.com/cool-boy-klay/go_gateway/dao"
	"github.com/cool-boy-klay/go_gateway/middleware"
	"github.com/cool-boy-klay/go_gateway/reverse_proxy"
	"github.com/gin-gonic/gin"
)

//匹配接入方式,基于请求信息
func HTTPReverseProxyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		//创建reverseProxy
		serverInterface,ok:=c.Get("service")
		if !ok{
			middleware.ResponseError(c,2001,errors.New("service not found"))
			//终止中间件向下传递
			c.Abort()
			return
		}
		serviceDetail:=serverInterface.(*dao.ServiceDetail)
		lb,err:=dao.LoadBalancerHandler.GetLoadBalancer(serviceDetail)
		if err!=nil{
			middleware.ResponseError(c,2002,err)
			//终止中间件向下传递
			c.Abort()
			return
		}
		trans,err:=dao.TransportHandler.GetTransporter(serviceDetail)
		if err!=nil{
			middleware.ResponseError(c,2003,err)
			//终止中间件向下传递
			c.Abort()
			return
		}
		proxy:=reverse_proxy.NewLoadBalanceReverseProxy(c,lb,trans)
		//使用reverseProxy.ServerHTTP(c.Request,c.Response)
		proxy.ServeHTTP(c.Writer,c.Request)


	}
}

