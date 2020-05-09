package gobom

import (
	"github.com/valyala/fasthttp"
	"gobom/utils"
	"time"
)

type Http struct {
	startTime  time.Duration
	endTime    time.Duration
	err        error
	request    *fasthttp.Request
	response   *fasthttp.Response
	resultResp chan<- *Response
}

func NewHttpRequest(respCh chan<- *Response) *Http {
	return &Http{
		request:    fasthttp.AcquireRequest(),
		response:   fasthttp.AcquireResponse(),
		resultResp: respCh,
	}
}

func (http *Http) send() {
	defer func() {
		fasthttp.ReleaseRequest(http.request)
		fasthttp.ReleaseResponse(http.response)
		http.recv()
	}()
	http.startTime = utils.Now()

	http.err = fasthttp.DoTimeout(http.request, http.response, time.Duration(5))
}

func (http *Http) recv() {
	IsSuccess := false
	if http.response.StatusCode() == fasthttp.StatusOK {
		IsSuccess = true
	}

	http.endTime = utils.Now()
	resp := &Response{
		WasteTime: uint64(http.getRequestTime()),
		IsSuccess: IsSuccess,
		ErrCode:   http.response.StatusCode(),
		ErrMsg:    http.err.Error(),
		Data:      nil,
	}
	http.resultResp <- resp
}

func (http *Http) getRequestTime() time.Duration {
	if http.startTime == 0 || http.endTime == 0 || http.endTime < http.startTime {
		return time.Duration(0)
	}
	return http.endTime - http.startTime
}
