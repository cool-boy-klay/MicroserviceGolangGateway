package main

import (
	"flag"
	"fmt"
	"github.com/cool-boy-klay/go_gateway/dao"
	"github.com/cool-boy-klay/go_gateway/http_proxy_router"
	"github.com/cool-boy-klay/go_gateway/router"
	"github.com/cool-boy-klay/go_gateway/tcp_proxy_router"
	"github.com/e421083458/golang_common/lib"
	"os"
	"os/signal"
	"syscall"
)

var (
	endpoint  =  flag.String("endpoint","","input endpoint dashboard or server")
)
func main()  {
	flag.Parse()
	if *endpoint==""{
		flag.Usage()
		os.Exit(1)
	}
	//if *config ==""{
	//	flag.Usage()
	//	os.Exit(1)
	//}

	if *endpoint =="dashboard"{
		lib.InitModule("./conf/dev/",[]string{"base","mysql","redis",})
		defer lib.Destroy()
		router.HttpServerRun()

		quit := make(chan os.Signal)
		signal.Notify(quit, syscall.SIGKILL, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		router.HttpServerStop()

	} else{
		lib.InitModule("./conf/dev/",[]string{"base","mysql","redis",})
		defer lib.Destroy()

		//从数据库中加载服务列表
		dao.ServiceManagerHandler.LoadOnce()
		//从数据库中加载租户列表
		dao.AppManagerHandler.LoadOnce()

		go func() {
			http_proxy_router.HttpServerRun()
		}()
		go func() {
			http_proxy_router.HttpsServerRun()
		}()
		go func() {
			tcp_proxy_router.TCPServerRun()
		}()


		fmt.Println("start server")
		quit := make(chan os.Signal)
		signal.Notify(quit, syscall.SIGKILL, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		http_proxy_router.HttpServerStop()
		http_proxy_router.HttpsServerStop()
		tcp_proxy_router.TCPServerStop()
	}



}