package middleware

import (
	"errors"
	"github.com/cool-boy-klay/go_gateway/public"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
)

func SessionAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		if name,ok:=session.Get(public.AdminSessionInfoKey).(string);!ok||name==""{
			ResponseError(c, InternalErrorCode, errors.New("user not login"))
			c.Abort()
			return
		}
		c.Next()
	}
}
