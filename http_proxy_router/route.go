package http_proxy_router

import (
	"github.com/cool-boy-klay/go_gateway/controller"
	"github.com/cool-boy-klay/go_gateway/http_proxy_middleware"
	"github.com/cool-boy-klay/go_gateway/middleware"
	"github.com/gin-gonic/gin"
)


func InitRouter(middlewares ...gin.HandlerFunc) *gin.Engine {

	router := gin.Default()
	router.Use(middlewares...)
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	//Oauth
	oauth:=router.Group("/oauth")
	oauth.Use(middleware.TranslationMiddleware())
	{
		controller.OAuthRegister(oauth)
	}


	router.Use(
		//验证请求host在反向代理中是否存在
		http_proxy_middleware.HttpAccessModeMiddleware(),
		//流量统计
		http_proxy_middleware.HttpFlowCountMiddleware(),
		//服务限流
		http_proxy_middleware.HTTPFlowLimitMiddleware(),
		//JWT租户权限认证
		http_proxy_middleware.HttpJwtAuthTokenMiddleware(),
		//JWT租户流量统计
		http_proxy_middleware.HttpJwtFlowCountMiddleware(),
		//JWT租户流量限制
		http_proxy_middleware.HTTPJwtFlowLimitMiddleware(),
		//白名单
		http_proxy_middleware.HttpWhiteListMiddleware(),
		//黑名单
		http_proxy_middleware.HttpBlackListMiddleware(),
		//Header头转换
		http_proxy_middleware.HttpHeaderTransferMiddleware(),
		//Uri转换
		http_proxy_middleware.HttpStripUriMiddleware(),
		//url重写
		http_proxy_middleware.HttpUrlRewriteMiddleware(),
		//代理转发
		http_proxy_middleware.HTTPReverseProxyMiddleware(),
		)




	return router
}
