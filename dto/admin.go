package dto

import (
	"github.com/cool-boy-klay/go_gateway/public"
	"github.com/gin-gonic/gin"
	"time"
)

//获取用户信息返回json数据
type AdminInfoOutput struct {
	ID int `json:"id"`
	UserName string `json:"user_name"`
	LoginTime time.Time `json:"login_time"`
	Avatar string `json:"avatar"`
	Introduction string `json:"introduction"`
	Roles []string `json:"roles"`
}

//修改密码参数
type AdminChangePasswordInput struct {
	PassWord string `form:"password" json:"password" comment:"密码" example:"123456" validate:"required"`
}
//判断参数是否出错
func (param *AdminChangePasswordInput) BindValidParam(c *gin.Context) error{
	return public.DefaultGetValidParams(c,param)
}