package dto

import (
	"github.com/cool-boy-klay/go_gateway/public"
	"github.com/gin-gonic/gin"
)



type TokensInput struct {
	GrantType string `form:"grant_type" json:"grant_type" comment:"授权类型" example:"client_credentials" validate:"required"`
	Scope string `form:"scope" json:"scope" comment:"权限范围" example:"read_write" validate:"required"`
}

//判断参数是否出错
func (param *TokensInput) BindValidParam(c *gin.Context) error{
	return public.DefaultGetValidParams(c,param)
}

type TokensOutput struct {
	AccessToken string `form:"access_token" json:"grant_type" `
	ExpiresIn int `form:"expires_in" json:"expires_in" `
	TokenType string `form:"token_type" json:"token_type" `
	Scope string `form:"scope" json:"scope" `

}
