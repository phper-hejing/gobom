package gobom

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/donnie4w/go-logger/logger"
)

type Requester interface {
	dispose() (response *Response, err error)
	send() (err error)
	recv() (response *Response, err error)
	close()
}

type GobomRequest struct {
	Options    *Options `json:"options"`
	Report     *Report  `json:"report"`
	Duration   *uint64  `json:"duration"`
	ConCurrent *uint64  `json:"conCurrent"`

	wg         sync.WaitGroup
	stop       chan bool
	stopStatus bool // 标识stop chan是否关闭
	resultResp chan *Response
}

type Response struct {
	WasteTime            uint64            `json:"wasteTime"`            // 消耗时间（毫秒）
	IsSuccess            bool              `json:"isSuccess"`            // 是否请求成功
	ErrCode              int               `json:"errCode"`              // 错误码
	ErrMsg               string            `json:"errMsg"`               // 错误提示
	Data                 []byte            `json:"report"`               // 响应数据
	TransactionWasteTime map[string]uint64 `json:"transactionWasteTime"` // 事务中每个步骤消耗时间
}

const (
	DEFAULT_STOP_CAP        = 1 << 20
	DEFAULT_RESPONSE_COUNT  = 1000
	DEFAULT_REQUEST_TIMEOUT = 5 // 连接超时（秒）
	ERR_RETRIES             = 3 // 失败重试次数
	CLOSE_ALL               = 0
)

type DisposeCallFunc func(err error) error

func NewGomBomRequest(options *Options) (*GobomRequest, error) {
	if err := options.Check(); err != nil {
		return nil, err
	}
	gobom := GobomRequest{
		Report:     &Report{},
		Options:    options,
		ConCurrent: new(uint64),
		Duration:   new(uint64),
	}
	atomic.StoreUint64(gobom.ConCurrent, options.ConCurrent)
	atomic.StoreUint64(gobom.Duration, options.Duration)
	return &gobom, nil
}

func (gobom *GobomRequest) Dispose(callback DisposeCallFunc) error {

	// 预请求
	//if err := gobom.boardTest(); err != nil {
	//	return err
	//}

	logger.Debug("dispose start...")

	var (
		ReportWg   sync.WaitGroup // 统计数据wg
		conCurrent = gobom.getConCurrent()
		duration   = gobom.getDuration()
		err        error
	)

	gobom.wg = sync.WaitGroup{}
	gobom.resultResp = make(chan *Response, DEFAULT_RESPONSE_COUNT)
	gobom.stop = make(chan bool, DEFAULT_STOP_CAP)

	go gobom.Timer() // 定时器关闭请求
	ReportWg.Add(1)
	go gobom.Report.ReceivingResults(gobom.resultResp, &ReportWg) // 统计请求数据

	gobom.Start(gobom.getConCurrent())

	gobom.wg.Wait()
	close(gobom.stop)
	close(gobom.resultResp)
	ReportWg.Wait()
	gobom.stopStatus = true
	logger.Debug("dispose out...")

	atomic.StoreUint64(gobom.ConCurrent, conCurrent)
	atomic.StoreUint64(gobom.Duration, duration)

	callback(err)

	return err
}

func (gobom *GobomRequest) AddConcurrentAndStart(count uint64) {
	gobom.addConCurrent(count)
	gobom.Start(count)
}

func (gobom *GobomRequest) Start(count uint64) {
	logger.Debug("signal count：", gobom.getConCurrent())
	for i := uint64(0); i < count; i++ {
		gobom.wg.Add(1)
		go func() {
			defer func() {
				gobom.wg.Done()
				gobom.minusConCurrent(1)
			}()
			if err := gobom.board(); err != nil {
				logger.Debug(err)
				gobom.PushResponse(&Response{
					IsSuccess: false,
					ErrCode:   -1,
					ErrMsg:    err.Error(),
				})
			}
		}()
	}
}

func (gobom *GobomRequest) boardTest() (err error) {
	requester, err := gobom.GetRequester()
	if err != nil {
		return err
	}
	defer requester.close()
	_, err = requester.dispose()
	if err != nil {
		return err
	}
	return nil
}

func (gobom *GobomRequest) board() (err error) {

	requester, err := gobom.GetRequester()
	if err != nil {
		return err
	}
	defer requester.close()

	err_retries := 1
	for {
		select {
		case <-gobom.stop:
			return nil
		default:
			resp, err := requester.dispose()
			if err != nil {
				if err_retries > ERR_RETRIES {
					return err
				}
				err_retries++
				continue
			} else {
				err_retries = 1
			}

			gobom.PushResponse(resp)

			if gobom.Options.Interval != 0 {
				time.Sleep(time.Duration(gobom.Options.Interval) * time.Millisecond)
			}
		}
	}
}

func (gobom *GobomRequest) Close(count uint64) {
	if gobom.stopStatus {
		return
	}
	if count == CLOSE_ALL || count >= gobom.getConCurrent() {
		count = gobom.getConCurrent()
	}
	var i uint64
	for i = 0; i < count; i++ {
		gobom.stop <- true
	}
	logger.Debug("close signal count: ", i)
	gobom.minusConCurrent(count)
	//for {
	//	// 外层逻辑需要确保压测逻辑全部结束在返回
	//	if gobom.stopStatus == true {
	//		return
	//	}
	//	time.Sleep(time.Duration(100) * time.Millisecond)
	//}
}

func (gobom *GobomRequest) Timer() {
	if gobom.getDuration() == 0 {
		return
	}
	t := time.NewTicker(1 * time.Second)
	for {
		if gobom.getDuration() == 0 {
			break
		}
		<-t.C
		gobom.minusDuration(1)
	}
	gobom.Close(CLOSE_ALL)
}

func (gobom *GobomRequest) Info() *Report {
	return gobom.Report
}

func (gobom *GobomRequest) getConCurrent() uint64 {
	return atomic.LoadUint64(gobom.ConCurrent)
}

func (gobom *GobomRequest) getDuration() uint64 {
	return atomic.LoadUint64(gobom.Duration)
}

func (gobom *GobomRequest) setConCurrent(n uint64) {
	atomic.StoreUint64(gobom.ConCurrent, n)
}

func (gobom *GobomRequest) setDuration(n uint64) {
	atomic.StoreUint64(gobom.Duration, n)
}

func (gobom *GobomRequest) addConCurrent(count uint64) {
	atomic.AddUint64(gobom.ConCurrent, count)
}

func (gobom *GobomRequest) minusConCurrent(count uint64) {
	if atomic.LoadUint64(gobom.ConCurrent) == 0 {
		return
	}
	atomic.AddUint64(gobom.ConCurrent, ^uint64(count-1))
}

func (gobom *GobomRequest) addDuration(count uint64) {
	atomic.AddUint64(gobom.Duration, count)
}

func (gobom *GobomRequest) minusDuration(count uint64) {
	if atomic.LoadUint64(gobom.Duration) == 0 {
		return
	}
	atomic.AddUint64(gobom.Duration, ^uint64(count-1))
}

func (gobom *GobomRequest) PushResponse(resp *Response) {
	if resp != nil {
		gobom.resultResp <- resp
	}
}

func (gobom *GobomRequest) GetRequester() (requester Requester, err error) {
	switch gobom.Options.Form {
	case FORM_HTTP:
		requester, err = NewHttpRequest(gobom.Options)
	case FORM_TCP:
		requester, err = NewTcpRequest(gobom.Options)
	case FORM_WEBSOCKET:
		// TODO
	default:
		return nil, ERR_FORM
	}
	if err != nil {
		return nil, err
	}
	return requester, nil
}
