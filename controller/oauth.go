package controller

import (
	"encoding/base64"
	"fmt"
	"github.com/cool-boy-klay/go_gateway/dao"
	"github.com/cool-boy-klay/go_gateway/dto"
	"github.com/cool-boy-klay/go_gateway/middleware"
	"github.com/cool-boy-klay/go_gateway/public"
	"github.com/dgrijalva/jwt-go"
	"github.com/e421083458/golang_common/lib"
	"github.com/pkg/errors"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type OAuthController struct {

}

//注册方法
func OAuthRegister(group *gin.RouterGroup){
	oauth:=&OAuthController{}
	group.POST("/tokens",oauth.Tokens)

}



// Tokens godoc
// @Summary 获取TOKEN
// @Description 获取TOKEN
// @Tags OAUTH
// @ID /oauth/tokens
// @Accept json
// @Produce json
// @Param body body dto.TokensInput true "body"
// @Success 200 {object} middleware.Response{data=dto.TokensOutput} "success"
// @Router /oauth/tokens [post]
func (oauth *OAuthController)Tokens(c *gin.Context){
	params:=&dto.TokensInput{}
	if err:=params.BindValidParam(c);err!=nil{
		middleware.ResponseError(c,2000,err)
	}

	splits:=strings.Split(c.GetHeader("Authorization")," ")
	if len(splits)!=2{
		middleware.ResponseError(c,2001,errors.New("用户名或密码格式错误"))
		return
	}

	appSecret,err:=base64.StdEncoding.DecodeString(splits[1])
	if err!=nil{
		middleware.ResponseError(c,2003,err)
	}
	fmt.Println("appSecret:",string(appSecret))
	//appSecret: app_id_b:a561990s
	appSplit:=strings.Split(string(appSecret),":")
	if len(appSplit)!=2{
		middleware.ResponseError(c,2004,errors.New("用户名或密码格式错误"))
		return
	}

	appList:=dao.AppManagerHandler.GetAppList()
	for _,appInfo:=range appList{
		if appInfo.AppID==appSplit[0]&&appInfo.Secret==appSplit[1]{
			claims:=jwt.StandardClaims{
				Issuer:appInfo.AppID,
				//过期时间=当前时间+服务器设定的过期间隔
				ExpiresAt:time.Now().Add(public.JWtExpiresAt*time.Second).In(lib.TimeLocation).Unix(),
			}
			token,err:=public.JwtEncode(claims)
			if err!=nil{
				middleware.ResponseError(c,2005,err)
			}

			output:=&dto.TokensOutput{
				AccessToken: token,
				ExpiresIn:   public.JWtExpiresAt,
				TokenType:   "Bearer",
				Scope:       "read_write",
			}
			middleware.ResponseSuccess(c,output)
			return
		}
	}


	middleware.ResponseError(c,2006,errors.New("未匹配到正确app信息"))


}

