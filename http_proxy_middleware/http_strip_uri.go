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

//http头转换
func HttpStripUriMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		serverInterface,ok:=c.Get("service")
		if !ok{
			middleware.ResponseError(c,2001,errors.New("service not found"))
			c.Abort()
			return
		}
		serviceDetail:=serverInterface.(*dao.ServiceDetail)

		//http://127.0.0.1:7999/test_http_string/abbb
		//http://127.0.0.1:3004/abb
		if serviceDetail.Http.RuleType==public.HTTPRuleTypePrefixURL &&serviceDetail.Http.NeedStripUri==1{
			fmt.Println("c.Request.URL.Path:",c.Request.URL.Path)
			c.Request.URL.Path = strings.Replace(c.Request.URL.Path,serviceDetail.Http.Rule,"",1)
			fmt.Println("c.Request.URL.Path:",c.Request.URL.Path)

		}

		c.Next()
	}
}
