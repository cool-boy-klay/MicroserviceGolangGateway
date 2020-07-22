package dao

import (
	"errors"
	"fmt"
	"github.com/cool-boy-klay/go_gateway/dto"
	"github.com/cool-boy-klay/go_gateway/public"
	"github.com/e421083458/golang_common/lib"
	"github.com/gin-gonic/gin"
	"net/http/httptest"
	"strings"
	"sync"
)

type ServiceDetail struct {
	Info *ServiceInfo `json:"info" description:"基本信息"`
	Http *HttpRule `json:"http" description:"http_rule"`
	TCP *TcpRule `json:"tcp" description:"tcp_rule"`
	GRPC *GrpcRule `json:"grpc" description:"grpc_rule"`
	LoadBalance *LoadBalance `json:"load_balance" description:"load_balance"`
	AccessControl *AccessControl `json:"access_control" description:"access_control"`
}

var ServiceManagerHandler *ServiceManager

func init(){
	ServiceManagerHandler = NewServiceManage()
}

type ServiceManager struct {
	ServiceMap map[string]*ServiceDetail
	ServiceSlice []*ServiceDetail
	Locker sync.RWMutex
	init sync.Once
	err error
}

func NewServiceManage() *ServiceManager{
	return &ServiceManager{
		ServiceMap: map[string]*ServiceDetail{},
		ServiceSlice:[]*ServiceDetail{},
		Locker:sync.RWMutex{},
		init:sync.Once{},
	}
}

func (s*ServiceManager) GetTcpServiceList() []*ServiceDetail{
	var list []*ServiceDetail
	for _,serviceItem :=range s.ServiceSlice{
		tempItem:=serviceItem
		if tempItem.Info.LoadType==public.LoadTypeTCP{
			list = append(list,tempItem)
		}
	}
	return list
}

func (s*ServiceManager) GetGrpcServiceList() []*ServiceDetail{
	var list []*ServiceDetail
	for _,serviceItem :=range s.ServiceSlice{
		tempItem:=serviceItem
		if tempItem.Info.LoadType==public.LoadTypeGRPC{
			list = append(list,tempItem)
		}
	}
	return list
}



func (s *ServiceManager) HTTPAccessMode(c *gin.Context) (*ServiceDetail,error){
	//1.前缀匹配 /abc => serviceSlice.rule
	//2.域名匹配 www.test.com => serviceSlice.rule

	//host c.Request.Host
	//path c.Request.URL.Path


	host:=c.Request.Host //www.test.com:8080
	host=host[0:strings.Index(host,":")]
	fmt.Println("host",host)

	path:=c.Request.URL.Path //   /abc/get



	for _,serviceItem:=range s.ServiceSlice{
		if serviceItem.Info.LoadType !=public.LoadTypeHTTP{
			continue
		}
		//域名
		if serviceItem.Http.RuleType == public.HTTPRuleTypeDomain{
			if serviceItem.Http.Rule == host{
				return serviceItem,nil
			}
		}
		if serviceItem.Http.RuleType == public.HTTPRuleTypePrefixURL{
			if strings.HasPrefix(path,serviceItem.Http.Rule){
				return serviceItem,nil
			}
		}

	}


	return nil,errors.New("not matched service")
}


func (s *ServiceManager) LoadOnce() error{
	s.init.Do(func() {

		serviceInfo:=&ServiceInfo{}
		c,_:=gin.CreateTestContext(httptest.NewRecorder())
		tx,err:=lib.GetGormPool("default")
		if err!=nil{
			s.err = err
			return
		}
		params:=&dto.ServiceListInput{PageNo:1,PageSize:99999}
		list,_,err:=serviceInfo.PageList(c,tx,params)
		if err!=nil{
			s.err = err
			return
		}
		s.Locker.Lock()
		defer s.Locker.Unlock()
		for _,listItem:=range list{
			tempItem:=listItem
			serviceDetail,err:=listItem.ServiceDetail(c,tx,&tempItem)
			if err!=nil{
				s.err = err
				return
			}
			s.ServiceMap[listItem.ServiceName] = serviceDetail
			s.ServiceSlice = append(s.ServiceSlice,serviceDetail)

		}
	})

	return s.err
}