package public

import (
	"golang.org/x/time/rate"
	"sync"
)

var FlowLimitHandler *FlowLimiter

type FlowLimiter struct {
	RedisFlowLimitMap map[string]*FlowLimiterItem
	RedisFlowLimitSlice []*FlowLimiterItem
	Locker sync.RWMutex

}

type FlowLimiterItem struct {
	ServiceName string
	Limiter *rate.Limiter
}



func NewFlowLimiter() *FlowLimiter{

	return &FlowLimiter{
		RedisFlowLimitMap:   map[string]*FlowLimiterItem{},
		RedisFlowLimitSlice: []*FlowLimiterItem{},
		Locker:sync.RWMutex{},
	}

}

func init(){
	FlowLimitHandler = NewFlowLimiter()

}

func (counter *FlowLimiter)GetFlowLimit(serviceName string,qps float64) (*rate.Limiter,error){

	for _,item:=range counter.RedisFlowLimitSlice {
		if item.ServiceName ==serviceName{
			return item.Limiter,nil
		}
	}
	//每秒token数，最大token数
	newLimiter:=rate.NewLimiter(rate.Limit(qps),int(qps*3))

	newLimiterItem:=&FlowLimiterItem{
		Limiter:newLimiter,
		ServiceName:serviceName,
	}

	counter.RedisFlowLimitSlice = append(counter.RedisFlowLimitSlice,newLimiterItem)

	counter.Locker.Lock()
	defer counter.Locker.Unlock()
	counter.RedisFlowLimitMap[serviceName] = newLimiterItem
	return newLimiterItem.Limiter,nil

}
