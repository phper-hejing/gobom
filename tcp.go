package gobom

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/smallnest/goframe"

	"gobom/utils"
)

type Tcp struct {
	startTime          time.Duration
	endTime            time.Duration
	err                error
	errRetries         uint8
	resultResp         chan<- *Response
	opt                *Options
	frameConn          goframe.FrameConn
	TransactionOptions *TransactionOptions
}

type TcpPools struct {
	connPools map[string]chan goframe.FrameConn
	mu        sync.RWMutex
}

var tcoPools = &TcpPools{
	connPools: make(map[string]chan goframe.FrameConn),
	mu:        sync.RWMutex{},
}

const (
	TYPE_NONE = iota
	TYPE_LINEBASEDFRAMECODEC
	TYPE_DELIMITERBASEDFRAMECODEC
	TYPE_FIXEDLENGTHFRAMECODEC
	TYPE_LENGTHFIELDBASEDFRAMECODEC
)

func (tcpPools *TcpPools) get(name, url string) (frameConn goframe.FrameConn, err error) {
	tcpPools.mu.Lock()
	connChan, ok := tcpPools.connPools[name]
	tcpPools.mu.Unlock()
	if !ok {
		tcpPools.mu.Lock()
		tcpPools.connPools[name] = make(chan goframe.FrameConn, 1024)
		tcpPools.mu.Unlock()
		frameConn, err = newFrameConn(url)
	} else {
		select {
		case frameConn = <-connChan:
		case <-time.After(5 * time.Second):
			frameConn, err = newFrameConn(url)
		}
	}
	return frameConn, err
}

func (tcpPools *TcpPools) put(name string, frameConn goframe.FrameConn) {
	tcpPools.mu.Lock()
	defer tcpPools.mu.Unlock()
	if connChan, ok := tcpPools.connPools[name]; ok {
		connChan <- frameConn
	}
}

func newFrameConn(url string) (frameConn goframe.FrameConn, err error) {
	conn, err := net.Dial("tcp", url)
	if err != nil {
		return nil, err
	}
	encoderConfig := goframe.EncoderConfig{
		ByteOrder:                       binary.BigEndian,
		LengthFieldLength:               4,
		LengthAdjustment:                0,
		LengthIncludesLengthFieldLength: false,
	}
	decoderConfig := goframe.DecoderConfig{
		ByteOrder:           binary.BigEndian,
		LengthFieldOffset:   0,
		LengthFieldLength:   4,
		LengthAdjustment:    0,
		InitialBytesToStrip: 4,
	}
	return goframe.NewLengthFieldBasedFrameConn(encoderConfig, decoderConfig, conn), nil
}

func NewTcpRequest(opt *Options) (*Tcp, error) {

	tcp := &Tcp{}
	tcp.opt = opt
	tcp.errRetries = ERR_RETRIES
	tcp.TransactionOptions = &TransactionOptions{
		TransactionOptionsData: opt.TransactionOptions.TransactionOptionsData,
		TransactionResponse:    make(map[string][]byte),
		TransactionIndex:       opt.TransactionOptions.TransactionIndex,
	}
	return tcp, nil
}

func (tcp *Tcp) dispose() (response *Response, err error) {

	if !tcp.TransactionOptions.Empty() {
		response := &Response{
			TransactionWasteTime: make(map[string]uint64),
		}
		respTemp := &Response{}
		isSuccess := true
		for _, data := range tcp.TransactionOptions.TransactionOptionsData {
			if err = tcp.send(); err != nil {
				err = fmt.Errorf(fmt.Sprint(data.Name, "，错误原因：", err.Error()))
				isSuccess = false
				break
			}
			respTemp, err = tcp.recv()
			if err != nil {
				err = fmt.Errorf(fmt.Sprint(data.Name, "，错误原因：", err.Error()))
				isSuccess = false
				break
			} else {
				tcoPools.put(data.Name, tcp.frameConn)
			}
			if respTemp.Data != nil {
				tcp.TransactionOptions.SetTransactionResponse(data.Name, respTemp.Data)
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

	if err = tcp.send(); err != nil {
		return nil, err
	}
	response, err = tcp.recv()
	if err == nil {
		tcoPools.put(tcp.opt.Url, tcp.frameConn)
	}
	return
}

func (tcp *Tcp) send() (err error) {
	var frameConn goframe.FrameConn
	var transactionData TransactionOptionsData
	if tcp.TransactionOptions != nil && tcp.TransactionOptions.TransactionOptionsData != nil {
		transactionData = tcp.TransactionOptions.Get()
		transactionData.SendData.init()
		frameConn, err = tcoPools.get(transactionData.Name, transactionData.Url)
	} else {
		frameConn, err = tcoPools.get(tcp.opt.Url, tcp.opt.Url)
	}
	if err != nil {
		return err
	}
	frameConn, err = newFrameConn(tcp.opt.Url)

	tcp.startTime = utils.Now()
	frameConn.Conn().SetReadDeadline(time.Now().Add(DEFAULT_REQUEST_TIMEOUT * time.Second))
	frameConn.Conn().SetWriteDeadline(time.Now().Add(DEFAULT_REQUEST_TIMEOUT * time.Second))
	if err := frameConn.WriteFrame(tcp.getSendData(transactionData)); err != nil {
		return err
	}
	tcp.frameConn = frameConn

	return nil
}

func (tcp *Tcp) recv() (response *Response, err error) {
	isSuccess := true
	errMsg := ""
	data := make([]byte, 0)

	data, err = tcp.frameConn.ReadFrame()
	tcp.endTime = utils.Now()
	if err != nil {
		isSuccess = false
		errMsg = err.Error()
	}
	return &Response{
		WasteTime: uint64(tcp.getRequestTime()),
		IsSuccess: isSuccess,
		ErrMsg:    errMsg,
		Data:      data,
	}, err
}

func (tcp *Tcp) close() {

}

func (tcp *Tcp) getRequestTime() time.Duration {
	if tcp.startTime == 0 || tcp.endTime == 0 || tcp.endTime < tcp.startTime {
		return time.Duration(0)
	}
	return tcp.endTime - tcp.startTime
}

func (tcp *Tcp) getSendData(transactionData TransactionOptionsData) []byte {
	//dataByte, _ := proto.Marshal(&protocol.User{
	//	Name: "testUser",
	//	Age:  13,
	//	Sex:  1,
	//})
	dataByte, _ := json.Marshal(transactionData.SendData.GetSendDataToMap(tcp.TransactionOptions))
	dataByte = []byte("aaa")
	return dataByte
}
