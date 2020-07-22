package dao

import (
	"fmt"
	"github.com/cool-boy-klay/go_gateway/public"
	"github.com/cool-boy-klay/go_gateway/reverse_proxy/load_balance"
	"github.com/e421083458/gorm"
	"github.com/gin-gonic/gin"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type LoadBalance struct {
	ID            int64  `json:"id" gorm:"primary_key"`
	ServiceID     int64  `json:"service_id" gorm:"column:service_id" description:"服务id	"`
	CheckMethod   int    `json:"check_method" gorm:"column:check_method" description:"检查方法 tcpchk=检测端口是否握手成功	"`
	CheckTimeout  int    `json:"check_timeout" gorm:"column:check_timeout" description:"check超时时间	"`
	CheckInterval int    `json:"check_interval" gorm:"column:check_interval" description:"检查间隔, 单位s		"`
	RoundType     int    `json:"round_type" gorm:"column:round_type" description:"轮询方式 round/weight_round/random/ip_hash"`
	IpList        string `json:"ip_list" gorm:"column:ip_list" description:"ip列表"`
	WeightList    string `json:"weight_list" gorm:"column:weight_list" description:"权重列表"`
	ForbidList    string `json:"forbid_list" gorm:"column:forbid_list" description:"禁用ip列表"`

	UpstreamConnectTimeout int `json:"upstream_connect_timeout" gorm:"column:upstream_connect_timeout" description:"下游建立连接超时, 单位s"`
	UpstreamHeaderTimeout  int `json:"upstream_header_timeout" gorm:"column:upstream_header_timeout" description:"下游获取header超时, 单位s	"`
	UpstreamIdleTimeout    int `json:"upstream_idle_timeout" gorm:"column:upstream_idle_timeout" description:"下游链接最大空闲时间, 单位s	"`
	UpstreamMaxIdle        int `json:"upstream_max_idle" gorm:"column:upstream_max_idle" description:"下游最大空闲链接数"`
}

func (t *LoadBalance) TableName() string {
	return "gateway_service_load_balance"
}

func (t *LoadBalance) Find(c *gin.Context, tx *gorm.DB, search *LoadBalance) (*LoadBalance, error) {
	model := &LoadBalance{}
	err := tx.SetCtx(public.GetGinTraceContext(c)).Where(search).Find(model).Error
	return model, err
}

func (t *LoadBalance) Save(c *gin.Context, tx *gorm.DB) error {
	if err := tx.SetCtx(public.GetGinTraceContext(c)).Save(t).Error; err != nil {
		return err
	}
	return nil
}
func (t *LoadBalance) GetIpListMyModel() []string {
	return strings.Split(t.IpList,",")
}

func (t *LoadBalance) GetWeightListMyModel() []string {
	return strings.Split(t.WeightList,",")
}

var LoadBalancerHandler *LoadBalancer

type LoadBalancer struct {
	LoadBalanceMap map[string]*LoadBalancerItem
	LoadBalanceSlice []*LoadBalancerItem
	Locker sync.RWMutex

}

type LoadBalancerItem struct {
	LoadBalance load_balance.LoadBalance
	ServiceName string
}



func NewLoadBalancer() *LoadBalancer{
	return &LoadBalancer{
		LoadBalanceMap:   map[string]*LoadBalancerItem{},
		LoadBalanceSlice: []*LoadBalancerItem{},
		Locker:sync.RWMutex{},
	}

}

func init(){
	LoadBalancerHandler = NewLoadBalancer()


}

func (l *LoadBalancer)GetLoadBalancer(service *ServiceDetail) (load_balance.LoadBalance,error){

	for _,lbrItem:=range l.LoadBalanceSlice{
		if lbrItem.ServiceName==service.Info.ServiceName{
			return lbrItem.LoadBalance,nil
		}
	}



	schema:="http://"
	if service.Http.NeedHttps==1{
		schema="https://"
	}
	if service.Info.LoadType==public.LoadTypeTCP||service.Info.LoadType==public.LoadTypeGRPC{
		schema = ""
	}


	//prefix:=""
	//if service.Http.RuleType==public.HTTPRuleTypePrefixURL{
	//	prefix = service.Http.Rule
	//}

	ipList:=service.LoadBalance.GetIpListMyModel()
	weightList:=service.LoadBalance.GetWeightListMyModel()
	ipConf:=map[string]string{}
	for ipIndex,ipItem:=range ipList{
		ipConf[ipItem]=weightList[ipIndex]
	}
	mConf,err:=load_balance.NewLoadBalanceCheckConf(fmt.Sprintf("%s%s", schema, "%s"), ipConf)
	//mConf,err:=load_balance.NewLoadBalanceCheckConf(
	//	fmt.Sprintf("%s://%s%s",schema,prefix),
	//	ipConf,
	//	)


	if err!=nil{

		return nil,err
	}

	lb:= load_balance.LoadBanlanceFactorWithConf(load_balance.LbType(service.LoadBalance.RoundType),mConf)
	lbrItem:=&LoadBalancerItem{
		LoadBalance:       lb,
		ServiceName: service.Info.ServiceName,
	}

	l.LoadBalanceSlice = append(l.LoadBalanceSlice,lbrItem)
	//写map需要加锁
	l.Locker.Lock()
	defer l.Locker.Unlock()
	l.LoadBalanceMap[service.Info.ServiceName]= lbrItem

	return lbrItem.LoadBalance,nil

}



var TransportHandler *Transporter

type Transporter struct {
	TransportMap map[string]*TransporterItem
	TransportSlice []*TransporterItem
	Locker sync.RWMutex

}
type TransporterItem struct {
	trans *http.Transport
	ServiceName string

}


func NewTransporter() *Transporter{
	return &Transporter{
		TransportMap:   map[string]*TransporterItem{},
		TransportSlice: []*TransporterItem{},
		Locker:sync.RWMutex{},
	}

}

func init(){
	TransportHandler = NewTransporter()


}

func (t *Transporter)GetTransporter(service *ServiceDetail) (*http.Transport,error){
	for _,transItem:=range t.TransportSlice{
		if transItem.ServiceName==service.Info.ServiceName{
			return transItem.trans,nil
		}

	}


	trans:=&http.Transport{
		DialContext:            (&net.Dialer{
			Timeout: time.Duration(service.LoadBalance.UpstreamConnectTimeout)*time.Second,

		}).DialContext,
		MaxIdleConns:service.LoadBalance.UpstreamMaxIdle,
		IdleConnTimeout:time.Duration(service.LoadBalance.UpstreamIdleTimeout),
		ResponseHeaderTimeout:time.Duration(service.LoadBalance.UpstreamHeaderTimeout),
	}

	transItem:=&TransporterItem{
		trans:       trans,
		ServiceName: service.Info.ServiceName,
	}

	t.TransportSlice = append(t.TransportSlice,transItem)
	//写map需要加锁
	t.Locker.Lock()
	defer t.Locker.Unlock()
	t.TransportMap[service.Info.ServiceName] = transItem
	return trans,nil

}