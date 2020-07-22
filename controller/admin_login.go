package controller

import (
	"encoding/json"
	"github.com/cool-boy-klay/go_gateway/dao"
	"github.com/cool-boy-klay/go_gateway/dto"
	"github.com/cool-boy-klay/go_gateway/middleware"
	"github.com/cool-boy-klay/go_gateway/public"
	"github.com/e421083458/golang_common/lib"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"time"
)

type AdminLoginController struct {

}

//注册方法
func AdminLoginRegister(group *gin.RouterGroup){
	admin:=&AdminLoginController{}
	group.POST("/login",admin.AdminLogin)
	group.GET("/logout",admin.AdminLogout)
	group.OPTIONS("/login",admin.AdminOption)
}


func (adminLogin *AdminLoginController)AdminOption(c *gin.Context){

	middleware.ResponseSuccess(c,nil)


}



// AdminLogin godoc
// @Summary 管理员登陆
// @Description 管理员登陆
// @Tags 管理员接口
// @ID /admin_login/login
// @Accept json
// @Produce json
// @Param body body dto.AdminLoginInput true "body"
// @Success 200 {object} middleware.Response{data=dto.AdminLoginOutput} "success"
// @Router /admin_login/login [post]
func (adminLogin *AdminLoginController)AdminLogin(c *gin.Context){
	params:=&dto.AdminLoginInput{}
	if err:=params.BindValidParam(c);err!=nil{
		middleware.ResponseError(c,2000,err)
	}

	//获取数据库连接池
	tx,err:=lib.GetGormPool("default")
	if err!=nil{
		middleware.ResponseError(c,2001,err)
	}

	//从数据库中获取username对应的用户信息并验证密码是否正确
	admin:=&dao.Admin{}
	admin,err=admin.LoginCheck(c,tx,params)
	if err!=nil{
		middleware.ResponseError(c,2002,err)
	}

	//	设置session
	sessInfo:=&dto.AdminSessionInfo{
		ID:admin.Id,
		UserName:admin.UserName,
		LoginTime:time.Now(),
	}
	sessBts,err:=json.Marshal(sessInfo)
	if err!=nil{
		middleware.ResponseError(c,2003,err)
	}

	sess:=sessions.Default(c)
	sess.Set(public.AdminSessionInfoKey,string(sessBts))
	sess.Save()

	out:=&dto.AdminLoginOutput{Token:admin.UserName}
	middleware.ResponseSuccess(c,out)


}


// AdminLogout godoc
// @Summary 管理员退出登陆
// @Description 管理员退出登陆
// @Tags 管理员接口
// @ID /admin_login/logout
// @Accept json
// @Produce json
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /admin_login/logout [get]
func (adminLogin *AdminLoginController)AdminLogout(c *gin.Context){
	//删除对应的session
	sess:=sessions.Default(c)
	sess.Delete(public.AdminSessionInfoKey)
	if err:=sess.Save();err!=nil{
		middleware.ResponseError(c,2005,err)
	}
	middleware.ResponseSuccess(c,"登出成功")

}
