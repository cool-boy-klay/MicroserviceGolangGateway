package controller

import (
	"encoding/json"
	"fmt"
	"github.com/cool-boy-klay/go_gateway/dao"
	"github.com/cool-boy-klay/go_gateway/dto"
	"github.com/cool-boy-klay/go_gateway/middleware"
	"github.com/cool-boy-klay/go_gateway/public"
	"github.com/e421083458/golang_common/lib"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
)

type AdminController struct {

}

//注册方法
func AdminRegister(group *gin.RouterGroup){
	admin:=&AdminController{}
	group.GET("/admin_info",admin.AdminInfo)
	group.POST("/change_password",admin.AdminChangePassWord)
}



// Admin godoc
// @Summary 管理员信息
// @Description 管理员信息
// @Tags 管理员接口
// @ID /admin/admin_info
// @Accept json
// @Produce json
// @Success 200 {object} middleware.Response{data=dto.AdminInfoOutput} "success"
// @Router /admin/admin_info [get]
func (adminLogin *AdminController)AdminInfo(c *gin.Context){
	//读取session信息
	sess:=sessions.Default(c)
	sessInfo:=sess.Get(public.AdminSessionInfoKey)
	adminSessionInfo:=&dto.AdminSessionInfo{}
	//把session信息转换为结构体
	if err:=json.Unmarshal([]byte(fmt.Sprint(sessInfo)),adminSessionInfo);err!=nil{
		middleware.ResponseError(c,2004,err)
		return
	}
	out:=&dto.AdminInfoOutput{
		ID:           adminSessionInfo.ID,
		UserName:     adminSessionInfo.UserName,
		LoginTime:    adminSessionInfo.LoginTime,
		Avatar:       "https://wpimg.wallstcn.com/f778738c-e4f8-4870-b634-56703b4acafe.gif",
		Introduction: "Super administrator",
		Roles:        []string{"admin"},
	}
	middleware.ResponseSuccess(c,out)
}


// AdminChangePassWord godoc
// @Summary 管理员修改密码
// @Description 管理员修改密码
// @Tags 管理员接口
// @ID /admin/change_password
// @Accept json
// @Produce json
// @Param body body dto.AdminChangePasswordInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /admin/change_password [post]
func (adminLogin *AdminController)AdminChangePassWord(c *gin.Context){

	//1.绑定参数结构体
	params:=&dto.AdminChangePasswordInput{}
	if err:=params.BindValidParam(c);err!=nil{
		middleware.ResponseError(c,2007,err)
	}

	//2.session读取信息到结构体
	//读取session信息
	sess:=sessions.Default(c)
	sessInfo:=sess.Get(public.AdminSessionInfoKey)
	adminSessionInfo:=&dto.AdminSessionInfo{}
	//把session信息转换为结构体
	if err:=json.Unmarshal([]byte(fmt.Sprint(sessInfo)),adminSessionInfo);err!=nil{
		middleware.ResponseError(c,2008,err)
		return
	}
	//3.根据session得到ID，利用ID去数据库中读取salt
	//获取数据库连接池
	tx,err:=lib.GetGormPool("default")
	if err!=nil{
		middleware.ResponseError(c,2009,err)
	}
	adminInfo:=&dao.Admin{}
	adminInfo,err=adminInfo.Find(c,tx,&dao.Admin{Id:adminSessionInfo.ID,IsDelete:0})
	if err!=nil{
		middleware.ResponseError(c,2010,err)
	}
	//生成新密码
	saltPassword :=public.GenSaltPassword(adminInfo.Salt,params.PassWord)
	adminInfo.Password=saltPassword
	//执行保存
	if err=adminInfo.Save(c,tx);err!=nil{
		middleware.ResponseError(c,2011,err)
	}

	middleware.ResponseSuccess(c,"修改密码成功")


}