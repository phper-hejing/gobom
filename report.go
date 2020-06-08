package gobom

import (
	"sync"
)

type Report struct {
	ConCurrency               uint64              `json:"conCurrency"` // 请求并发数
	TotalTime                 uint64              `json:"totalTime"`   // 任务处理总时间(成功请求)
	MaxTime                   uint64              `json:"maxTime"`     // 单个请求最大消耗时长(成功请求)
	MinTime                   uint64              `json:"minTime"`     // 单个请求最小消耗时长(成功请求)
	AverageTime               uint64              `json:"averageTime"` // 平均每个请求消耗时长(成功请求)
	SuccessNum                uint64              `json:"successNum"`  // 成功请求数
	FailureNum                uint64              `json:"failureNum"`  // 失败请求数
	ErrCode                   map[int]int         `json:"errCode"`     // [错误码]错误个数
	ErrCodeMsg                map[int]string      `json:"errCodeMsg"`  // [错误码]错误码描述
	EveryReqWasteTime         []uint64            `json:"-"`           // 每一个请求/事务 消耗的时间记录
	EveryTransactionWasteTime []map[string]uint64 `json:"-"`           // 每一个事务中的每个步骤消耗的时间记录
}

const FILE_STORE_PATH = "./store/report"

func (report *Report) ReceivingResults(resultResp <-chan *Response, ReportWg *sync.WaitGroup) {
	defer ReportWg.Done()

	var (
		totalTime                 uint64                 // 总时间
		maxTime                   uint64                 // 最大时长
		minTime                   uint64                 // 最小时长
		successNum                uint64                 // 成功请求数
		failureNum                uint64                 // 失败请求数
		errCode                   = make(map[int]int)    // [错误码]错误个数
		errCodeMsg                = make(map[int]string) // [错误码]错误码描述
		everyReqWasteTime         = make([]uint64, 0)
		everyTransactionWasteTime = make([]map[string]uint64, 0)
	)

	for data := range resultResp {
		if data.IsSuccess {
			totalTime += data.WasteTime
			successNum++
			if data.WasteTime > maxTime {
				maxTime = data.WasteTime
			}
			if minTime == 0 || (data.WasteTime != 0 && data.WasteTime < minTime) {
				minTime = data.WasteTime
			}
			everyReqWasteTime = append(everyReqWasteTime, data.WasteTime)
			if data.TransactionWasteTime != nil {
				everyTransactionWasteTime = append(everyTransactionWasteTime, data.TransactionWasteTime)
			}
		} else {
			failureNum++
			errCode[data.ErrCode]++
			if _, ok := errCodeMsg[data.ErrCode]; !ok {
				errCodeMsg[data.ErrCode] = data.ErrMsg
			}
		}

		report.MaxTime = maxTime
		report.MinTime = minTime
		report.SuccessNum = successNum
		report.FailureNum = failureNum
		report.ErrCode = errCode
		report.ErrCodeMsg = errCodeMsg
		report.TotalTime = totalTime
		report.AverageTime = report.getAvgTime()
		report.EveryReqWasteTime = everyReqWasteTime
		report.EveryTransactionWasteTime = everyTransactionWasteTime
	}

}

func (report *Report) getAvgTime() uint64 {
	if report.TotalTime == 0 || report.SuccessNum == 0 {
		return 0
	} else {
		return report.TotalTime / report.SuccessNum
	}
}
