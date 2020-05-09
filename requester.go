package gobom

import (
	"errors"
	"sync"
	"time"

	"github.com/donnie4w/go-logger/logger"
)

type Requester interface {
	send()
	recv()
}

type GobomRequest struct {
	options    *Options
	request    Requester
	report     *Report
	resultResp chan *Response

	stop chan bool
	once sync.Once
}

type Response struct {
	WasteTime uint64      `json:"wasteTime"` // 消耗时间（毫秒）
	IsSuccess bool        `json:"isSuccess"` // 是否请求成功
	ErrCode   int         `json:"errCode"`   // 错误码
	ErrMsg    string      `json:"errMsg"`    // 错误提示
	Data      interface{} `json:"data"`      // 响应数据
}

const (
	DEFAULT_RESPONSE_COUNT = 1000
)

var (
	ErrConCurrency = errors.New("并发数不能小于1")
)

func NewGomBomRequest(id string, options Options) *GobomRequest {
	resultResp := make(chan *Response, DEFAULT_RESPONSE_COUNT)
	gobom := GobomRequest{
		report: &Report{
			id:          id,
			ConCurrency: options.ConCurrency,
		},
		resultResp: resultResp,
	}
	switch options.Form {
	case FORM_HTTP:
		gobom.request = NewHttpRequest(resultResp)
	case FORM_TCP:
		// TODO
	case FORM_WEBSOCKET:
		// TODO
	default:
		return nil
	}
	return &gobom
}

func (gobom GobomRequest) Dispose(id string) error {
	if err := gobom.check(); err != nil {
		return err
	}

	var (
		ReqWg    sync.WaitGroup // 并发请求wg
		ReportWg sync.WaitGroup // 统计数据wg
	)

	go gobom.Timer() // 定时器关闭请求
	ReportWg.Add(1)
	go gobom.report.ReceivingResults(gobom.resultResp, &ReportWg) // 统计请求数据

	gobom.stop = make(chan bool, gobom.getConCurrency()) // 关闭chan

	for i := uint64(0); i < gobom.getConCurrency(); i++ {
		ReqWg.Add(1)
		go func() {
			defer ReqWg.Done()
			for {
				select {
				case <-gobom.stop:
					return
				default:
					gobom.request.send()
				}
			}
		}()
	}

	ReqWg.Wait()
	close(gobom.resultResp)
	close(gobom.stop)
	ReportWg.Wait()
	logger.Debug("dispose out...")

	return nil
}

func (gobom GobomRequest) Close() {
	gobom.once.Do(func() {
		for i := uint64(0); i < gobom.getConCurrency(); i++ {
			gobom.stop <- true
			logger.Debug("close signal number: ", i)
		}
	})
}

func (gobom GobomRequest) Timer() {
	if gobom.getDuration() == 0 {
		return
	}
	t := time.NewTicker(1 * time.Second)
	for {
		if gobom.options.Duration == 0 {
			break
		}
		<-t.C
		gobom.options.Duration--
	}
	gobom.Close()
}

func (gobom GobomRequest) getConCurrency() uint64 {
	return gobom.options.ConCurrency
}

func (gobom GobomRequest) getForm() int {
	return gobom.options.Form
}

func (gobom GobomRequest) getInterval() uint64 {
	return gobom.options.Interval
}

func (gobom GobomRequest) getDuration() uint64 {
	return gobom.options.Duration
}

func (gobom GobomRequest) check() error {
	if gobom.getConCurrency() <= 0 {
		return ErrConCurrency
	}
	return nil
}
