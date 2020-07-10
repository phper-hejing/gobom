package gobom

import (
	"sync"
)

type Report struct {
	TaskId                    string              `json:"taskId"`
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

	if report.ErrCode == nil {
		report.ErrCode = make(map[int]int)
	}

	if report.ErrCodeMsg == nil {
		report.ErrCodeMsg = make(map[int]string)
	}

	if report.EveryReqWasteTime == nil {
		report.EveryReqWasteTime = make([]uint64, 0)
	}

	if report.EveryTransactionWasteTime == nil {
		report.EveryTransactionWasteTime = make([]map[string]uint64, 0)
	}

	for data := range resultResp {
		if data.IsSuccess {
			report.TotalTime += data.WasteTime
			report.SuccessNum++
			if data.WasteTime > report.MaxTime {
				report.MaxTime = data.WasteTime
			}
			if report.MinTime == 0 || (data.WasteTime != 0 && data.WasteTime < report.MinTime) {
				report.MinTime = data.WasteTime
			}
			report.EveryReqWasteTime = append(report.EveryReqWasteTime, data.WasteTime)
			if data.TransactionWasteTime != nil {
				report.EveryTransactionWasteTime = append(report.EveryTransactionWasteTime, data.TransactionWasteTime)
			}
		} else {
			report.FailureNum++
			report.ErrCode[data.ErrCode]++
			if _, ok := report.ErrCodeMsg[data.ErrCode]; !ok {
				report.ErrCodeMsg[data.ErrCode] = data.ErrMsg
			}
		}
		report.AverageTime = report.getAvgTime()
	}

}

func (report *Report) getAvgTime() uint64 {
	if report.TotalTime == 0 || report.SuccessNum == 0 {
		return 0
	} else {
		return report.TotalTime / report.SuccessNum
	}
}
