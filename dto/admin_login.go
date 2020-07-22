package dto

import (
	"github.com/cool-boy-klay/go_gateway/public"
	"github.com/gin-gonic/gin"
	"time"
)
type AdminSessionInfo struct {
	ID int `json:"id"`
	UserName string `json:"user_name"`
	LoginTime time.Time `json:"login_time"`

}


type AdminLoginInput struct {
	UserName string `form:"username" json:"username" comment:"用户名" example:"admin" validate:"required,validate_username"`
	PassWord string `form:"password" json:"password" comment:"密码" example:"123456" validate:"required"`
}

//判断参数是否出错
func (param *AdminLoginInput) BindValidParam(c *gin.Context) error{
	return public.DefaultGetValidParams(c,param)
}

type AdminLoginOutput struct {
	Token string `form:"token" json:"token" comment:"token" example:"admin" validate:""`
}