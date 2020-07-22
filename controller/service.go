package controller

import (
	"fmt"
	"github.com/cool-boy-klay/go_gateway/dao"
	"github.com/cool-boy-klay/go_gateway/dto"
	"github.com/cool-boy-klay/go_gateway/middleware"
	"github.com/cool-boy-klay/go_gateway/public"
	"github.com/e421083458/golang_common/lib"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"strings"
	"time"
)

type ServiceController struct {

}

//注册方法
func ServiceRegister(group *gin.RouterGroup){
	service:=&ServiceController{}
	group.GET("/service_list",service.ServiceList)
	group.GET("/service_delete",service.ServiceDelete)
	group.POST("/service_add_http",service.ServiceAddHTTP)
	group.POST("/service_update_http", service.ServiceUpdateHTTP)
	group.GET("/service_detail",service.ServiceDetail)
	group.GET("/service_stat",service.ServiceStat)

	group.POST("/service_add_tcp", service.ServiceAddTcp)
	group.POST("/service_update_tcp", service.ServiceUpdateTcp)
	group.POST("/service_add_grpc", service.ServiceAddGrpc)
	group.POST("/service_update_grpc", service.ServiceUpdateGrpc)
}





// ServiceDelete godoc
// @Summary 删除特定的服务
// @Description 根据id删除特定的服务
// @Tags 服务管理
// @ID /service/service_delete
// @Accept json
// @Produce json
// @Param id query string true "id"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/service_delete [get]
func (adminLogin *ServiceController)ServiceDelete(c *gin.Context){

	//1.绑定参数结构体
	params:=&dto.ServiceDeleteInput{}
	if err:=params.BindValidParam(c);err!=nil{
		middleware.ResponseError(c,3005,err)
	}

	//获取数据库连接池
	tx,err:=lib.GetGormPool("default")
	if err!=nil{
		middleware.ResponseError(c,3001,err)
	}


	serviceInfo:=&dao.ServiceInfo{ID:params.ID}
	serviceInfo,err=serviceInfo.Find(c,tx,serviceInfo)
	if err!=nil{
		middleware.ResponseError(c,3006,err)
	}
	serviceInfo.IsDelete=1
	err=serviceInfo.Save(c,tx)
	if err!=nil{
		middleware.ResponseError(c,3007,err)
	}

	middleware.ResponseSuccess(c,"删除服务成功!")


}


// ServiceList godoc
// @Summary 获取服务列表
// @Description 获取服务列表
// @Tags 服务管理
// @ID /service/service_list
// @Accept json
// @Produce json
// @Param info query string false "关键词"
// @Param page_size query int true "每页个数"
// @Param page_no query int true "当前页数"
// @Success 200 {object} middleware.Response{data=dto.ServiceListOutput} "success"
// @Router /service/service_list [get]
func (adminLogin *ServiceController)ServiceList(c *gin.Context){

	//1.绑定参数结构体
	params:=&dto.ServiceListInput{}
	if err:=params.BindValidParam(c);err!=nil{
		middleware.ResponseError(c,3000,err)
	}

	//获取数据库连接池
	tx,err:=lib.GetGormPool("default")
	if err!=nil{
		middleware.ResponseError(c,3001,err)
	}

	serviceInfo:=&dao.ServiceInfo{}
	list,total,err:=serviceInfo.PageList(c,tx,params)
	if err!=nil{
		middleware.ResponseError(c,3002,err)
	}

	outList:=[]dto.ServiceListItemOutput{}
	for _,listItem:=range list{
		serviceDetail,err:=listItem.ServiceDetail(c,tx,&listItem)
		if err!=nil{
			middleware.ResponseError(c,3003,err)
			return
		}
		serviceAddr:="unKnow"
		clusterIP:=lib.GetStringConf("base.cluster.cluster_ip")
		clusterPort:=lib.GetStringConf("base.cluster.cluster_port")
		clusterSSLPort:=lib.GetStringConf("base.cluster.cluster_ssl_port")

		if serviceDetail.Info.LoadType==public.LoadTypeHTTP{

			if serviceDetail.Http.RuleType==public.HTTPRuleTypePrefixURL{
				if serviceDetail.Http.NeedHttps==1{
					serviceAddr = fmt.Sprintf("%s:%s%s",clusterIP,clusterSSLPort,serviceDetail.Http.Rule)

				}else{
					serviceAddr = fmt.Sprintf("%s:%s%s",clusterIP,clusterPort,serviceDetail.Http.Rule)

				}
			} else if serviceDetail.Http.RuleType==public.HTTPRuleTypeDomain{
				serviceAddr = serviceDetail.Http.Rule
			}

		}
		if serviceDetail.Info.LoadType==public.LoadTypeTCP{
			serviceAddr = fmt.Sprintf("%s:%d",clusterIP,serviceDetail.TCP.Port)
		}
		if serviceDetail.Info.LoadType==public.LoadTypeGRPC{
			serviceAddr = fmt.Sprintf("%s:%d",clusterIP,serviceDetail.GRPC.Port)
		}

		//利用LoadBalance获取ipList
		ipList:=serviceDetail.LoadBalance.GetIpListMyModel()

		//从redis中获取Counter值
		serviceCounter,err:=public.FlowCounterHandler.GetFlowCounter(public.FlowServicePrefix+listItem.ServiceName)
		if err!=nil{
			middleware.ResponseError(c,3004,err)
			return
		}

		outItem:=dto.ServiceListItemOutput{
			ID:listItem.ID,
			ServiceName:listItem.ServiceName,
			ServiceDesc:listItem.ServiceDesc,
			ServiceAddr:serviceAddr,
			Qps:serviceCounter.QPS,
			Qpd:serviceCounter.TotalCount,
			TotalNode:int64(len(ipList)),
		}
		outList = append(outList, outItem)

	}

	out:=&dto.ServiceListOutput{Total:total,List:outList}

	middleware.ResponseSuccess(c,out)


}



// ServiceAddHTTP godoc
// @Summary 添加HTTP服务
// @Description 添加HTTP服务
// @Tags 服务管理
// @ID /service/service_add_http
// @Accept  json
// @Produce  json
// @Param body body dto.ServiceAddHTTPInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/service_add_http [post]
func (service *ServiceController) ServiceAddHTTP(c *gin.Context) {
	params := &dto.ServiceAddHTTPInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 3000, err)
		return
	}

	tx,err:=lib.GetGormPool("default")
	if err!=nil{
		middleware.ResponseError(c,3001,err)
	}
	tx = tx.Begin()

	//从数据库中查找是否有相同名称的记录
	serviceInfo:=&dao.ServiceInfo{ServiceName:params.ServiceName}
	if _,err=serviceInfo.Find(c,tx,serviceInfo);err==nil{
		tx.Rollback()
		middleware.ResponseError(c,3004,errors.New("服务已经存在"))
		return
	}

	//从数据库中查找是否有相同ruleType和Rule的记录
	httpUrl:=&dao.HttpRule{RuleType:params.RuleType,Rule:params.Rule}
	if _,err=httpUrl.Find(c,tx,httpUrl);err==nil{
		tx.Rollback()
		middleware.ResponseError(c,3005,errors.New("服务接入前缀或者域名已经存在"))
		return
	}
	//判断权重列表个数与IP地址个数是否相同
	if len(strings.Split(params.IpList,"\n"))!=len(strings.Split(params.WeightList,"\n")){
		tx.Rollback()
		middleware.ResponseError(c,3005,errors.New("IP地址个数与权重列表个数不一致"))
		return
	}

	serviceModel := &dao.ServiceInfo{
		ServiceName: params.ServiceName,
		ServiceDesc: params.ServiceDesc,
		LoadType:public.LoadTypeHTTP,
	}
	if err := serviceModel.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 3007, err)
		return
	}
	//serviceModel.ID
	httpRule := &dao.HttpRule{
		ServiceID:      serviceModel.ID,
		RuleType:       params.RuleType,
		Rule:           params.Rule,
		NeedHttps:      params.NeedHttps,
		NeedStripUri:   params.NeedStripUri,
		NeedWebsocket:  params.NeedWebsocket,
		UrlRewrite:     params.UrlRewrite,
		HeaderTransfor: params.HeaderTransfor,
	}
	if err := httpRule.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 3008, err)
		return
	}

	accessControl := &dao.AccessControl{
		ServiceID:         serviceModel.ID,
		OpenAuth:          params.OpenAuth,
		BlackList:         params.BlackList,
		WhiteList:         params.WhiteList,
		ClientIPFlowLimit: params.ClientipFlowLimit,
		ServiceFlowLimit:  params.ServiceFlowLimit,
	}
	if err := accessControl.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 3009, err)
		return
	}

	loadbalance := &dao.LoadBalance{
		ServiceID:              serviceModel.ID,
		RoundType:              params.RoundType,
		IpList:                 params.IpList,
		WeightList:             params.WeightList,
		UpstreamConnectTimeout: params.UpstreamConnectTimeout,
		UpstreamHeaderTimeout:  params.UpstreamHeaderTimeout,
		UpstreamIdleTimeout:    params.UpstreamIdleTimeout,
		UpstreamMaxIdle:        params.UpstreamMaxIdle,
	}
	if err := loadbalance.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 3010, err)
		return
	}
	tx.Commit()
	middleware.ResponseSuccess(c, "新建HTTP服务成功")


}



// ServiceUpdateHTTP godoc
// @Summary 修改HTTP服务
// @Description 修改HTTP服务
// @Tags 服务管理
// @ID /service/service_update_http
// @Accept  json
// @Produce  json
// @Param body body dto.ServiceUpdateHTTPInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/service_update_http [post]
func (service *ServiceController) ServiceUpdateHTTP(c *gin.Context) {
	params := &dto.ServiceUpdateHTTPInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 2000, err)
		return
	}

	if len(strings.Split(params.IpList, ",")) != len(strings.Split(params.WeightList, ",")) {
		middleware.ResponseError(c, 2001, errors.New("IP列表与权重列表数量不一致"))
		return
	}

	tx, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}
	tx = tx.Begin()
	serviceInfo := &dao.ServiceInfo{ServiceName: params.ServiceName}
	serviceInfo, err = serviceInfo.Find(c, tx, serviceInfo)
	if err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2003, errors.New("服务不存在"))
		return
	}
	serviceDetail, err := serviceInfo.ServiceDetail(c, tx, serviceInfo)
	if err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2004, errors.New("服务不存在"))
		return
	}

	info := serviceDetail.Info
	info.ServiceDesc = params.ServiceDesc
	if err := info.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2005, err)
		return
	}

	httpRule := serviceDetail.Http
	httpRule.NeedHttps = params.NeedHttps
	httpRule.NeedStripUri = params.NeedStripUri
	httpRule.NeedWebsocket = params.NeedWebsocket
	httpRule.UrlRewrite = params.UrlRewrite
	httpRule.HeaderTransfor = params.HeaderTransfor
	if err := httpRule.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2006, err)
		return
	}

	accessControl := serviceDetail.AccessControl
	accessControl.OpenAuth = params.OpenAuth
	accessControl.BlackList = params.BlackList
	accessControl.WhiteList = params.WhiteList
	accessControl.ClientIPFlowLimit = params.ClientipFlowLimit
	accessControl.ServiceFlowLimit = params.ServiceFlowLimit
	if err := accessControl.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2007, err)
		return
	}

	loadbalance := serviceDetail.LoadBalance
	loadbalance.RoundType = params.RoundType
	loadbalance.IpList = params.IpList
	loadbalance.WeightList = params.WeightList
	loadbalance.UpstreamConnectTimeout = params.UpstreamConnectTimeout
	loadbalance.UpstreamHeaderTimeout = params.UpstreamHeaderTimeout
	loadbalance.UpstreamIdleTimeout = params.UpstreamIdleTimeout
	loadbalance.UpstreamMaxIdle = params.UpstreamMaxIdle
	if err := loadbalance.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2008, err)
		return
	}
	tx.Commit()
	middleware.ResponseSuccess(c, "")
}




// ServiceList godoc
// @Summary 获取服务详情
// @Description 获取服务详情信息
// @Tags 服务管理
// @ID /service/service_detail
// @Accept json
// @Produce json
// @Param id query string true "关键词"
// @Success 200 {object} middleware.Response{data=dao.ServiceDetail} "success"
// @Router /service/service_detail [get]
func (adminLogin *ServiceController)ServiceDetail(c *gin.Context){

	//1.绑定参数结构体
	params:=&dto.ServiceDetailInput{}
	if err:=params.BindValidParam(c);err!=nil{
		middleware.ResponseError(c,3000,err)
	}

	//获取数据库连接池
	tx,err:=lib.GetGormPool("default")
	if err!=nil{
		middleware.ResponseError(c,3001,err)
	}

	serviceInfo:=&dao.ServiceInfo{ID:params.ID}
	//根据对应ID查找serviceInfo的dao对象
	serviceInfo,err=serviceInfo.Find(c,tx,serviceInfo)
	if err!=nil{
		middleware.ResponseError(c,3002,err)
	}

	serviceDetail,err:=serviceInfo.ServiceDetail(c,tx,serviceInfo)
	if err!=nil{
		middleware.ResponseError(c,3002,err)
	}
	middleware.ResponseSuccess(c,serviceDetail)


}



// ServiceList godoc
// @Summary 服务统计
// @Description 根据服务id获取今天和昨天的数据
// @Tags 服务管理
// @ID /service/service_stat
// @Accept json
// @Produce json
// @Param id query string true "服务id"
// @Success 200 {object} middleware.Response{data=dto.ServiceStatOutput} "success"
// @Router /service/service_stat [get]
func (adminLogin *ServiceController)ServiceStat(c *gin.Context){

	//1.绑定参数结构体
	params:=&dto.ServiceStatInput{}
	if err:=params.BindValidParam(c);err!=nil{
		middleware.ResponseError(c,3000,err)
	}

	//获取数据库连接池
	tx,err:=lib.GetGormPool("default")
	if err!=nil{
		middleware.ResponseError(c,3001,err)
	}

	serviceInfo:=&dao.ServiceInfo{ID:params.ID}
	//根据对应ID查找serviceInfo的dao对象
	serviceInfo,err=serviceInfo.Find(c,tx,serviceInfo)
	if err!=nil{
		middleware.ResponseError(c,3002,err)
	}

	serviceDetail,err:=serviceInfo.ServiceDetail(c,tx,serviceInfo)
	if err!=nil{
		middleware.ResponseError(c,3002,err)
	}

	counter,err:=public.FlowCounterHandler.GetFlowCounter(public.FlowServicePrefix+serviceDetail.Info.ServiceName)
	if err!=nil{
		middleware.ResponseError(c,3003,err)
	}



	todayList:=[]int64{}
	currentTime := time.Now()
	for i:=0;i<=currentTime.Hour();i++{
		newTime:=time.Date(currentTime.Year(),currentTime.Month(),currentTime.Day(),i,0,0,0,lib.TimeLocation)
		hourDate,_:=counter.GetHourData(newTime)
		todayList = append(todayList,hourDate)
	}
	yesterdayList:=[]int64{}
	yesterdayTime:=currentTime.Add(-1*time.Duration(time.Hour*24))
	for i:=0;i<=23;i++{
		newTime:=time.Date(yesterdayTime.Year(),yesterdayTime.Month(),yesterdayTime.Day(),i,0,0,0,lib.TimeLocation)
		hourDate,_:=counter.GetHourData(newTime)
		yesterdayList = append(yesterdayList,hourDate)
	}

	out:=&dto.ServiceStatOutput{Today:todayList,Yesterday:yesterdayList}

	middleware.ResponseSuccess(c,out)


}



// ServiceAddHttp godoc
// @Summary tcp服务添加
// @Description tcp服务添加
// @Tags 服务管理
// @ID /service/service_add_tcp
// @Accept  json
// @Produce  json
// @Param body body dto.ServiceAddTcpInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/service_add_tcp [post]
func (admin *ServiceController) ServiceAddTcp(c *gin.Context) {
	params := &dto.ServiceAddTcpInput{}
	if err := params.GetValidParams(c); err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}

	//验证 service_name 是否被占用
	infoSearch := &dao.ServiceInfo{
		ServiceName: params.ServiceName,
		IsDelete:    0,
	}
	if _, err := infoSearch.Find(c, lib.GORMDefaultPool, infoSearch); err == nil {
		middleware.ResponseError(c, 2002, errors.New("服务名被占用，请重新输入"))
		return
	}

	//验证端口是否被占用?
	tcpRuleSearch := &dao.TcpRule{
		Port: params.Port,
	}
	if _, err := tcpRuleSearch.Find(c, lib.GORMDefaultPool, tcpRuleSearch); err == nil {
		middleware.ResponseError(c, 2003, errors.New("服务端口被占用，请重新输入"))
		return
	}
	grpcRuleSearch := &dao.GrpcRule{
		Port: params.Port,
	}
	if _, err := grpcRuleSearch.Find(c, lib.GORMDefaultPool, grpcRuleSearch); err == nil {
		middleware.ResponseError(c, 2004, errors.New("服务端口被占用，请重新输入"))
		return
	}

	//ip与权重数量一致
	if len(strings.Split(params.IpList, ",")) != len(strings.Split(params.WeightList, ",")) {
		middleware.ResponseError(c, 2005, errors.New("ip列表与权重设置不匹配"))
		return
	}

	tx := lib.GORMDefaultPool.Begin()
	info := &dao.ServiceInfo{
		LoadType:    public.LoadTypeTCP,
		ServiceName: params.ServiceName,
		ServiceDesc: params.ServiceDesc,
	}
	if err := info.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2006, err)
		return
	}
	loadBalance := &dao.LoadBalance{
		ServiceID:  info.ID,
		RoundType:  params.RoundType,
		IpList:     params.IpList,
		WeightList: params.WeightList,
		ForbidList: params.ForbidList,
	}
	if err := loadBalance.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2007, err)
		return
	}

	httpRule := &dao.TcpRule{
		ServiceID: info.ID,
		Port:      params.Port,
	}
	if err := httpRule.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2008, err)
		return
	}

	accessControl := &dao.AccessControl{
		ServiceID:         info.ID,
		OpenAuth:          params.OpenAuth,
		BlackList:         params.BlackList,
		WhiteList:         params.WhiteList,
		WhiteHostName:     params.WhiteHostName,
		ClientIPFlowLimit: params.ClientIPFlowLimit,
		ServiceFlowLimit:  params.ServiceFlowLimit,
	}
	if err := accessControl.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2009, err)
		return
	}
	tx.Commit()
	middleware.ResponseSuccess(c, "")
	return
}

// ServiceUpdateTcp godoc
// @Summary tcp服务更新
// @Description tcp服务更新
// @Tags 服务管理
// @ID /service/service_update_tcp
// @Accept  json
// @Produce  json
// @Param body body dto.ServiceUpdateTcpInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/service_update_tcp [post]
func (admin *ServiceController) ServiceUpdateTcp(c *gin.Context) {
	params := &dto.ServiceUpdateTcpInput{}
	if err := params.GetValidParams(c); err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}

	//ip与权重数量一致
	if len(strings.Split(params.IpList, ",")) != len(strings.Split(params.WeightList, ",")) {
		middleware.ResponseError(c, 2002, errors.New("ip列表与权重设置不匹配"))
		return
	}

	tx := lib.GORMDefaultPool.Begin()

	service := &dao.ServiceInfo{
		ID: params.ID,
	}
	detail, err := service.ServiceDetail(c, lib.GORMDefaultPool, service)
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}

	info := detail.Info
	info.ServiceDesc = params.ServiceDesc
	if err := info.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2003, err)
		return
	}

	loadBalance := &dao.LoadBalance{}
	if detail.LoadBalance != nil {
		loadBalance = detail.LoadBalance
	}
	loadBalance.ServiceID = info.ID
	loadBalance.RoundType = params.RoundType
	loadBalance.IpList = params.IpList
	loadBalance.WeightList = params.WeightList
	loadBalance.ForbidList = params.ForbidList
	if err := loadBalance.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2004, err)
		return
	}

	tcpRule := &dao.TcpRule{}
	if detail.TCP != nil {
		tcpRule = detail.TCP
	}
	tcpRule.ServiceID = info.ID
	tcpRule.Port = params.Port
	if err := tcpRule.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2005, err)
		return
	}

	accessControl := &dao.AccessControl{}
	if detail.AccessControl != nil {
		accessControl = detail.AccessControl
	}
	accessControl.ServiceID = info.ID
	accessControl.OpenAuth = params.OpenAuth
	accessControl.BlackList = params.BlackList
	accessControl.WhiteList = params.WhiteList
	accessControl.WhiteHostName = params.WhiteHostName
	accessControl.ClientIPFlowLimit = params.ClientIPFlowLimit
	accessControl.ServiceFlowLimit = params.ServiceFlowLimit
	if err := accessControl.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2006, err)
		return
	}
	tx.Commit()
	middleware.ResponseSuccess(c, "")
	return
}

// ServiceAddHttp godoc
// @Summary grpc服务添加
// @Description grpc服务添加
// @Tags 服务管理
// @ID /service/service_add_grpc
// @Accept  json
// @Produce  json
// @Param body body dto.ServiceAddGrpcInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/service_add_grpc [post]
func (admin *ServiceController) ServiceAddGrpc(c *gin.Context) {
	params := &dto.ServiceAddGrpcInput{}
	if err := params.GetValidParams(c); err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}

	//验证 service_name 是否被占用
	infoSearch := &dao.ServiceInfo{
		ServiceName: params.ServiceName,
		IsDelete:    0,
	}
	if _, err := infoSearch.Find(c, lib.GORMDefaultPool, infoSearch); err == nil {
		middleware.ResponseError(c, 2002, errors.New("服务名被占用，请重新输入"))
		return
	}

	//验证端口是否被占用?
	tcpRuleSearch := &dao.TcpRule{
		Port: params.Port,
	}
	if _, err := tcpRuleSearch.Find(c, lib.GORMDefaultPool, tcpRuleSearch); err == nil {
		middleware.ResponseError(c, 2003, errors.New("服务端口被占用，请重新输入"))
		return
	}
	grpcRuleSearch := &dao.GrpcRule{
		Port: params.Port,
	}
	if _, err := grpcRuleSearch.Find(c, lib.GORMDefaultPool, grpcRuleSearch); err == nil {
		middleware.ResponseError(c, 2004, errors.New("服务端口被占用，请重新输入"))
		return
	}

	//ip与权重数量一致
	if len(strings.Split(params.IpList, ",")) != len(strings.Split(params.WeightList, ",")) {
		middleware.ResponseError(c, 2005, errors.New("ip列表与权重设置不匹配"))
		return
	}

	tx := lib.GORMDefaultPool.Begin()
	info := &dao.ServiceInfo{
		LoadType:    public.LoadTypeGRPC,
		ServiceName: params.ServiceName,
		ServiceDesc: params.ServiceDesc,
	}
	if err := info.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2006, err)
		return
	}

	loadBalance := &dao.LoadBalance{
		ServiceID:  info.ID,
		RoundType:  params.RoundType,
		IpList:     params.IpList,
		WeightList: params.WeightList,
		ForbidList: params.ForbidList,
	}
	if err := loadBalance.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2007, err)
		return
	}

	grpcRule := &dao.GrpcRule{
		ServiceID:      info.ID,
		Port:           params.Port,
		HeaderTransfor: params.HeaderTransfor,
	}
	if err := grpcRule.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2008, err)
		return
	}

	accessControl := &dao.AccessControl{
		ServiceID:         info.ID,
		OpenAuth:          params.OpenAuth,
		BlackList:         params.BlackList,
		WhiteList:         params.WhiteList,
		WhiteHostName:     params.WhiteHostName,
		ClientIPFlowLimit: params.ClientIPFlowLimit,
		ServiceFlowLimit:  params.ServiceFlowLimit,
	}
	if err := accessControl.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2009, err)
		return
	}
	tx.Commit()
	middleware.ResponseSuccess(c, "")
	return
}

// ServiceUpdateTcp godoc
// @Summary grpc服务更新
// @Description grpc服务更新
// @Tags 服务管理
// @ID /service/service_update_grpc
// @Accept  json
// @Produce  json
// @Param body body dto.ServiceUpdateGrpcInput true "body"
// @Success 200 {object} middleware.Response{data=string} "success"
// @Router /service/service_update_grpc [post]
func (admin *ServiceController) ServiceUpdateGrpc(c *gin.Context) {
	params := &dto.ServiceUpdateGrpcInput{}
	if err := params.GetValidParams(c); err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}

	//ip与权重数量一致
	if len(strings.Split(params.IpList, ",")) != len(strings.Split(params.WeightList, ",")) {
		middleware.ResponseError(c, 2002, errors.New("ip列表与权重设置不匹配"))
		return
	}

	tx := lib.GORMDefaultPool.Begin()

	service := &dao.ServiceInfo{
		ID: params.ID,
	}
	detail, err := service.ServiceDetail(c, lib.GORMDefaultPool, service)
	if err != nil {
		middleware.ResponseError(c, 2003, err)
		return
	}

	info := detail.Info
	info.ServiceDesc = params.ServiceDesc
	if err := info.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2004, err)
		return
	}

	loadBalance := &dao.LoadBalance{}
	if detail.LoadBalance != nil {
		loadBalance = detail.LoadBalance
	}
	loadBalance.ServiceID = info.ID
	loadBalance.RoundType = params.RoundType
	loadBalance.IpList = params.IpList
	loadBalance.WeightList = params.WeightList
	loadBalance.ForbidList = params.ForbidList
	if err := loadBalance.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2005, err)
		return
	}

	grpcRule := &dao.GrpcRule{}
	if detail.GRPC != nil {
		grpcRule = detail.GRPC
	}
	grpcRule.ServiceID = info.ID
	//grpcRule.Port = params.Port
	grpcRule.HeaderTransfor = params.HeaderTransfor
	if err := grpcRule.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2006, err)
		return
	}

	accessControl := &dao.AccessControl{}
	if detail.AccessControl != nil {
		accessControl = detail.AccessControl
	}
	accessControl.ServiceID = info.ID
	accessControl.OpenAuth = params.OpenAuth
	accessControl.BlackList = params.BlackList
	accessControl.WhiteList = params.WhiteList
	accessControl.WhiteHostName = params.WhiteHostName
	accessControl.ClientIPFlowLimit = params.ClientIPFlowLimit
	accessControl.ServiceFlowLimit = params.ServiceFlowLimit
	if err := accessControl.Save(c, tx); err != nil {
		tx.Rollback()
		middleware.ResponseError(c, 2007, err)
		return
	}
	tx.Commit()
	middleware.ResponseSuccess(c, "")
	return
}
