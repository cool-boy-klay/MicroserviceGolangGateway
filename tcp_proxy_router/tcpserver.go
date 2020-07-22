package tcp_proxy_router

import (
	"context"
	"fmt"
	"github.com/cool-boy-klay/go_gateway/dao"
	"github.com/cool-boy-klay/go_gateway/tcp_proxy_middleware"
	"github.com/cool-boy-klay/go_gateway/tcp_server"


	"github.com/cool-boy-klay/go_gateway/reverse_proxy"
	"log"
	"net"
)
var tcpServerList = []*tcp_server.TcpServer{}


type tapHandler struct {

}
func (t *tapHandler) ServeTCP(ctx context.Context,src net.Conn){
	src.Write([]byte("TCPHandler\n"))
}

func TCPServerRun(){
	ServerList:=dao.ServiceManagerHandler.GetTcpServiceList()
	for _,serviceItem:=range ServerList{
		tempItem:=serviceItem
		log.Printf(" [INFO] TCPProxyRun:%v\n",tempItem.TCP.Port)

		go func(serviceDetail *dao.ServiceDetail) {
			addr := fmt.Sprintf(":%d", tempItem.TCP.Port)

			rb,err:=dao.LoadBalancerHandler.GetLoadBalancer(serviceDetail)
			if err!=nil{
				log.Fatalf(" [ERROR] GetTCPLoadBalancer:%v err:%v\n", serviceDetail.TCP.Port, err)
				return
			}

			//路由中间件
			router:=tcp_proxy_middleware.NewTcpSliceRouter()
			router.Group("/").Use(
				tcp_proxy_middleware.TCPFlowCountMiddleware(),
				tcp_proxy_middleware.TCPFlowLimitMiddleware(),
				tcp_proxy_middleware.TCPWhiteListMiddleware(),
				tcp_proxy_middleware.TCPBlackListMiddleware(),

			)
			//构建回调handler
			routerHandler:=tcp_proxy_middleware.NewTcpSliceRouterHandler(
					func(c *tcp_proxy_middleware.TcpSliceRouterContext) tcp_server.TCPHandler{
						return reverse_proxy.NewTcpLoadBalanceReverseProxy(c,rb)
					},router)

			baseCtx := context.WithValue(context.Background(),"service",serviceDetail)
			tcpServer:=&tcp_server.TcpServer{
				Addr:addr,
				Handler:routerHandler,
				BaseCtx:baseCtx,
			}
			tcpServerList = append(tcpServerList,tcpServer)

			err=tcpServer.ListenAndServe()
			if err!=nil && err!=tcp_server.ErrServerClosed{
				log.Fatalf(" [ERROR] TCPProxyRun:%v err:%v\n", serviceDetail.TCP.Port, err)

			}
		}(tempItem)

	}


}

func TCPServerStop(){
	for _,tcpServer:=range tcpServerList{
		tcpServer.Close()
		log.Printf(" [INFO] TCPProxyStop %v stopped\n",tcpServer.Addr)
	}

}
