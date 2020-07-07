package gobom

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/valyala/fasthttp"

	"gobom/utils"
)

type Http struct {
	startTime          time.Duration
	endTime            time.Duration
	err                error
	errRetries         uint8
	resultResp         chan<- *Response
	opt                *Options
	response           *fasthttp.Response
	TransactionOptions *TransactionOptions
}

func NewHttpRequest(opt *Options) (*Http, error) {
	return &Http{
		errRetries:         ERR_RETRIES,
		opt:                opt,
		TransactionOptions: opt.TransactionOptions.Copy(),
	}, nil
}

func (http *Http) dispose() (response *Response, err error) {
	if !http.TransactionOptions.Empty() {
		response := &Response{
			TransactionWasteTime: make(map[string]uint64),
		}
		respTemp := &Response{}
		isSuccess := true
		for _, data := range http.TransactionOptions.TransactionOptionsDataList {
			if err = http.send(); err != nil {
				err = fmt.Errorf(fmt.Sprint(data.Name, "，错误原因：", err.Error()))
				isSuccess = false
				break
			}
			respTemp, err = http.recv()
			if err != nil {
				err = fmt.Errorf(fmt.Sprint(data.Name, "，错误原因：", err.Error()))
				isSuccess = false
				break
			}
			if respTemp.Data != nil {
				http.TransactionOptions.SetTransactionResponse(data.Name, respTemp.Data)
			}
			if data.Interval != 0 {
				time.Sleep(time.Duration(data.Interval) * time.Millisecond)
			}
			response.TransactionWasteTime[data.Name] = respTemp.WasteTime
			response.WasteTime += respTemp.WasteTime
		}
		response.IsSuccess = isSuccess
		if err != nil {
			response.ErrMsg = err.Error()
		}

		return response, err
	}
	if err := http.send(); err != nil {
		return nil, err
	}
	return http.recv()
}

func (http *Http) send() (err error) {

	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()

	http.fillHttp(req, http.TransactionOptions)

	defer func() {
		fasthttp.ReleaseRequest(req)
	}()

	http.startTime = utils.Now()
	http.err = fasthttp.DoTimeout(req, resp, time.Duration(DEFAULT_REQUEST_TIMEOUT)*time.Second)
	http.response = resp

	return nil
}

func (http *Http) recv() (response *Response, err error) {

	defer func() {
		fasthttp.ReleaseResponse(http.response)
	}()

	isSuccess := true
	errMsg := ""
	if http.response.StatusCode() != fasthttp.StatusOK {
		isSuccess = false
		http.err = errors.New(fmt.Sprintf("错误码:%d", http.response.StatusCode()))
	}
	if http.err != nil {
		errMsg = http.err.Error()
	}
	http.endTime = utils.Now()

	return &Response{
		WasteTime: uint64(http.getRequestTime()),
		IsSuccess: isSuccess,
		ErrCode:   http.response.StatusCode(),
		ErrMsg:    errMsg,
		Data:      http.response.Body(),
	}, http.err
}

func (http *Http) close() {}

func (http *Http) getRequestTime() time.Duration {
	if http.startTime == 0 || http.endTime == 0 || http.endTime < http.startTime {
		return time.Duration(0)
	}
	return http.endTime - http.startTime
}

func (http *Http) fillHttp(req *fasthttp.Request, transactionOptions *TransactionOptions) {
	var (
		url      = http.opt.Url
		method   = http.opt.HttpOptions.Method
		cookie   = http.opt.HttpOptions.Cookie
		header   = http.opt.HttpOptions.Header
		sendData = http.opt.SendData
	)

	transactionOptionsData := transactionOptions.Get()
	if !transactionOptionsData.Empty() {
		sendData = transactionOptionsData.SendData
		url = transactionOptionsData.Url
		method = transactionOptionsData.HttpOptions.Method
		cookie = transactionOptionsData.HttpOptions.Cookie
		header = transactionOptionsData.HttpOptions.Header
	}

	if method == "" {
		method = "GET"
	}

	sendData.init()

	req.SetRequestURI(url)
	req.Header.SetMethod(method)
	req.Header.Set("user-agent", "gobom")
	req.Header.Set("Content-Type", "application/json")
	for k, v := range cookie {
		req.Header.SetCookie(k, v)
	}
	for k, v := range header {
		req.Header.Set(k, v)
	}
	if sendData != nil {
		if bm, err := json.Marshal(sendData.GetSendDataToMap(transactionOptions)); err == nil {
			req.SetBody(bm)
			if !transactionOptionsData.Empty() {
				transactionOptions.SetTransactionSendData(transactionOptionsData.Name, bm)
			}
		}
	}
}
