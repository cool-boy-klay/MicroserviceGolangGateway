package dao

import (
	"github.com/cool-boy-klay/go_gateway/dto"
	"github.com/cool-boy-klay/go_gateway/public"
	"github.com/e421083458/gorm"
	"github.com/gin-gonic/gin"
	"time"
)

type ServiceInfo struct {
	ID int64 `json:"id" gorm:"primary_key"`
	ServiceName string `json:"service_name" gorm:"column:service_name" description:"服务名称"` //服务名称
	ServiceDesc string `json:"service_desc" gorm:"column:service_desc" description:"服务描述"` //服务描述
	LoadType int `json:"load_type" gorm:"column:load_type" description:"负载类型 0=http 1=tcp 2=grpc"`			//服务类型
	CreatedAt time.Time `json:"create_at" gorm:"column:create_at" description:"添加时间"`
	UpdatedAt time.Time `json:"update_at" gorm:"column:update_at" description:"更新时间"`
	IsDelete int8 `json:"is_delete" gorm:"column:is_delete" description:"是否已经删除；0:否；1:是"`
}
func (t *ServiceInfo) TableName() string {
	return "gateway_service_info"
}

func (t *ServiceInfo) ServiceDetail(c *gin.Context, tx *gorm.DB, search *ServiceInfo) (*ServiceDetail,error) {

	httpRule:=&HttpRule{ServiceID:search.ID}
	httpRule,err:=httpRule.Find(c,tx,httpRule)
	if err!=nil&&err!=gorm.ErrRecordNotFound{
		return nil,err
	}

	tcpRule:=&TcpRule{ServiceID:search.ID}
	tcpRule,err=tcpRule.Find(c,tx,tcpRule)
	if err!=nil&&err!=gorm.ErrRecordNotFound{
		return nil,err
	}

	grpcRule:=&GrpcRule{ServiceID:search.ID}
	grpcRule,err=grpcRule.Find(c,tx,grpcRule)
	if err!=nil&&err!=gorm.ErrRecordNotFound{
		return nil,err
	}

	accessControl:=&AccessControl{ServiceID:search.ID}
	accessControl,err=accessControl.Find(c,tx,accessControl)
	if err!=nil&&err!=gorm.ErrRecordNotFound{
		return nil,err
	}

	loadBalance:=&LoadBalance{ServiceID:search.ID}
	loadBalance,err=loadBalance.Find(c,tx,loadBalance)
	if err!=nil&&err!=gorm.ErrRecordNotFound{
		return nil,err
	}

	detail:=&ServiceDetail{
		Info:          search,
		Http:          httpRule,
		TCP:           tcpRule,
		GRPC:          grpcRule,
		LoadBalance:   loadBalance,
		AccessControl: accessControl,
	}
	return detail,nil

}





//根据参数从数据库中查询服务列表
func (t *ServiceInfo) PageList(c *gin.Context, tx *gorm.DB, param *dto.ServiceListInput) ([]ServiceInfo, int64,error) {
	total:=int64(0)
	list:=[]ServiceInfo{}
	//在数据库表中的偏移量
	offset:=(param.PageNo-1)*param.PageSize
	query:=tx.SetCtx(public.GetGinTraceContext(c))
	query=query.Table(t.TableName()).Where("is_delete=0")

	if param.Info!=""{
		query=query.Where("(service_name like ? or service_desc like ?)","%"+param.Info+"%","%"+param.Info+"%")
	}
	if err:=query.Limit(param.PageSize).Offset(offset).Find(&list).Order("id desc").Error;err!=nil&&err!=gorm.ErrRecordNotFound{
		return nil,0,err
	}
	query.Limit(param.PageSize).Offset(offset).Count(&total)
	return list,total,nil
}


func (t *ServiceInfo) Find(c *gin.Context, tx *gorm.DB, search *ServiceInfo) (*ServiceInfo, error) {
	out:=&ServiceInfo{}
	err := tx.SetCtx(public.GetGinTraceContext(c)).Where(search).Find(out).Error
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (t *ServiceInfo) Save(c *gin.Context, tx *gorm.DB) error {

	return tx.SetCtx(public.GetGinTraceContext(c)).Save(t).Error

}


func (t *ServiceInfo) GroupByLoadType(c *gin.Context, tx *gorm.DB) ([]dto.DashServiceStatItemOutput, error) {
	list := []dto.DashServiceStatItemOutput{}
	query := tx.SetCtx(public.GetGinTraceContext(c))
	if err := query.Table(t.TableName()).Where("is_delete=0").Select("load_type, count(*) as value").Group("load_type").Scan(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}