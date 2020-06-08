package gobom

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tidwall/gjson"
	"strings"
	"sync"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/donnie4w/go-logger/logger"
	"github.com/smallnest/goframe"
	"github.com/valyala/fasthttp"

	"gobom/utils"
)

var (
	ERR_FILE_INIT  = errors.New("初始化失败")
	ERR_FILE_PARSE = errors.New("解析文件数据失败")
	ERR_FILE_OPEN  = errors.New("打开文件数据失败")
	ERR_FILE_READ  = errors.New("读取文件数据失败")
)

const (
	FORM_HTTP = iota
	FORM_TCP
	FORM_WEBSOCKET

	TYPE_INT    = "int"
	TYPE_STRING = "string"
	TYPE_FILE   = "file"
	TYPE_RESP   = "response"

	FILE_DATA_PATH = "../store/data"
	FILE_PARSE_SEP = "---"
)

type Options struct {
	TaskId           string `json:"taskId" form:"taskId"`                     // 任务id,（运行，删除）任务时使用
	Url              string `json:"url" form:"url"`                           // 请求地址
	ConCurrent       uint64 `json:"conCurrent" form:"conCurrent"`             // 并发数
	LessenConCurrent uint64 `json:"lessenConCurrent" form:"lessenConCurrent"` // 并发数（负数）
	Duration         uint64 `json:"duration" form:"duration"`                 // 持续时间（秒）
	Interval         uint64 `json:"interval" form:"interval"`                 // 请求间隔时间（毫秒）
	Form             int    `json:"form" form:"form"`                         // http|websocket|tcp

	SendData           *SendData          `json:"sendData"` // 压测数据
	HttpOptions        HttpOptions        `json:"httpOptions" form:"httpOptions"`
	TcpOptions         TcpOptions         `json:"tcpOptions" form:"tcpOptions"`
	TransactionOptions TransactionOptions `json:"transactionOptions" form:"transactionOptions"`
}

type TcpOptions struct {
	CodecType     uint `json:"codecType" form:"codecType"`
	encoderConfig goframe.EncoderConfig
	decoderConfig goframe.DecoderConfig
}

type HttpOptions struct {
	Method string            `json:"method" form:"method"` // 请求方法
	Cookie map[string]string `json:"cookie" form:"cookie"`
	Header map[string]string `json:"header" form:"header"`
}

type TransactionOptions struct {
	TransactionOptionsData []TransactionOptionsData `json:"transactionOptionsData"`
	TransactionResponse    map[string][]byte        `json:"-"`
	TransactionIndex       uint64                   `json:"-"`
}

type TransactionOptionsData struct {
	Name        string      `json:"name"`
	Url         string      `json:"url" form:"url"`           // 请求地址
	Interval    uint64      `json:"interval" form:"interval"` // 请求间隔时间（毫秒）
	HttpOptions HttpOptions `json:"httpOptions" form:"httpOptions"`
	SendData    *SendData   `json:"sendData"` // 压测数据
}

type SendData struct {
	DataFieldList []*DataField           `json:"dataFieldList" form:"dataFieldList"`
	SourceFileMap map[string]*SourceFile `json:"-"`
}

type DataField struct {
	Name    string      `json:"name" form:"name"`       // 字段名
	Type    string      `json:"type" form:"type"`       // 字段类型 int|string
	Len     int64       `json:"len" form:"len"`         // 字段长度 如果字段是int表示len中的随机数
	Default interface{} `json:"default" form:"default"` // 默认值
	Dynamic string      `json:"dynamic" form:"dynamic"` // 动态字段名（字段值从文件或其他请求响应中获取）
}

type SourceFile struct {
	Index  uint64     `json:"index"`
	Column []string   `json:"column"`
	Data   [][]string `json:"report"`
	m      sync.Mutex
}

var (
	ERR_URL         = errors.New("URL不能为空")
	ERR_CONCURRENCY = errors.New("并发数数值过小")
	ERR_DURATION    = errors.New("持续时间数值过小")
)

// 初始化数据
func (opt *Options) Init() {

}

func (opt *Options) Check() error {
	if opt.Url == "" {
		return ERR_URL
	}
	if opt.ConCurrent == 0 {
		return ERR_CONCURRENCY
	}
	return nil
}

func (opt *Options) fillHttp(req *fasthttp.Request, transactionOptions *TransactionOptions) {
	var (
		url      = opt.Url
		method   = opt.HttpOptions.Method
		cookie   = opt.HttpOptions.Cookie
		header   = opt.HttpOptions.Header
		sendData = opt.SendData
	)

	if transactionOptions != nil {
		if transactionOptions.TransactionOptionsData != nil && len(transactionOptions.TransactionOptionsData) != 0 {
			sendData = transactionOptions.TransactionOptionsData[transactionOptions.TransactionIndex].SendData
			url = transactionOptions.TransactionOptionsData[transactionOptions.TransactionIndex].Url
			method = transactionOptions.TransactionOptionsData[transactionOptions.TransactionIndex].HttpOptions.Method
			cookie = transactionOptions.TransactionOptionsData[transactionOptions.TransactionIndex].HttpOptions.Cookie
			header = transactionOptions.TransactionOptionsData[transactionOptions.TransactionIndex].HttpOptions.Header
		}
		defer func() {
			if transactionOptions.TransactionIndex++; transactionOptions.TransactionIndex >= uint64(len(transactionOptions.TransactionOptionsData)) {
				transactionOptions.TransactionIndex = 0
			}
		}()
	}

	if method == "" {
		method = "GET"
	}

	sendData.init()

	req.SetRequestURI(url)
	req.Header.SetMethod(method)
	req.Header.Set("user-agent", "gobom")
	for k, v := range cookie {
		req.Header.SetCookie(k, v)
	}
	for k, v := range header {
		req.Header.Set(k, v)
	}
	if sendData != nil {
		if bm, err := json.Marshal(sendData.GetSendDataToMap(transactionOptions)); err == nil {
			req.SetBody(bm)
		}
	}
}

func (sendData *SendData) init() error {
	if sendData == nil || sendData.DataFieldList == nil || sendData.SourceFileMap != nil {
		return nil
	}
	sendData.SourceFileMap = make(map[string]*SourceFile)
	for _, v := range sendData.DataFieldList {
		if v.Dynamic != "" && v.Type == TYPE_FILE {
			fileName, field := sendData.parseField(v.Dynamic)
			f, err := excelize.OpenFile(fmt.Sprintf("%s/%s", FILE_DATA_PATH, fileName))
			if err != nil {
				logger.Debug(err)
				return err
			}
			sheet := f.GetSheetMap()[1] // 获取excel的sheet名称
			csvData := f.GetRows(sheet)
			sendData.SourceFileMap[field] = &SourceFile{
				Index:  0,
				Column: csvData[0],
				Data:   csvData[1:],
			}
		}
	}
	return nil
}

func (tcpOptions *TcpOptions) init() error {
	if tcpOptions.CodecType == TYPE_NONE {
		tcpOptions.CodecType = TYPE_LENGTHFIELDBASEDFRAMECODEC
	}
	if tcpOptions.encoderConfig == (goframe.EncoderConfig{}) {
		tcpOptions.encoderConfig = goframe.EncoderConfig{
			ByteOrder:                       binary.BigEndian,
			LengthFieldLength:               4,
			LengthAdjustment:                0,
			LengthIncludesLengthFieldLength: false,
		}
	}
	if tcpOptions.decoderConfig == (goframe.DecoderConfig{}) {
		tcpOptions.decoderConfig = goframe.DecoderConfig{
			ByteOrder:           binary.BigEndian,
			LengthFieldOffset:   0,
			LengthFieldLength:   4,
			LengthAdjustment:    0,
			InitialBytesToStrip: 4,
		}
	}
	return nil
}

func (sendData *SendData) GetSendDataToMap(transactionOptions *TransactionOptions) map[string]interface{} {
	if sendData == nil || sendData.DataFieldList == nil {
		return nil
	}
	bm := make(map[string]interface{})
	for _, v := range sendData.DataFieldList {
		if v.Default != nil && v.Default != "" {
			bm[v.Name] = v.Default
			continue
		}
		switch v.Type {
		case TYPE_INT:
			bm[v.Name] = utils.GetRandomIntRange(uint64(v.Len))
		case TYPE_STRING:
			bm[v.Name] = utils.GetRandomStrings(uint64(v.Len))
		case TYPE_FILE:
			bm[v.Name] = sendData.getFileValue(v.Dynamic)
		case TYPE_RESP:
			bm[v.Name] = ""
			if transactionOptions != nil {
				name, fields := sendData.parseField(v.Dynamic)
				respByte := transactionOptions.GetTransactionResponse(name)
				if respByte == nil {
					continue
				}
				result := gjson.Get(string(respByte), fields)
				bm[v.Name] = result.Value()
			}
		}
	}
	return bm
}

func (transactionOptions *TransactionOptions) Get() TransactionOptionsData {
	defer func() {
		if transactionOptions.TransactionIndex++; transactionOptions.TransactionIndex >= uint64(len(transactionOptions.TransactionOptionsData)) {
			transactionOptions.TransactionIndex = 0
		}
	}()
	return transactionOptions.TransactionOptionsData[transactionOptions.TransactionIndex]
}

func (transactionOptions *TransactionOptions) GetTransactionResponse(name string) []byte {
	if transactionOptions == nil || transactionOptions.TransactionResponse == nil {
		return nil
	}
	if val, ok := transactionOptions.TransactionResponse[name]; ok {
		return val
	}
	return nil
}

func (transactionOptions *TransactionOptions) SetTransactionResponse(key string, val []byte) {
	if transactionOptions == nil || transactionOptions.TransactionResponse == nil {
		return
	}
	transactionOptions.TransactionResponse[key] = val
}

func (sendData *SendData) getFileValue(key string) interface{} {
	_, field := sendData.parseField(key)
	data, ok := sendData.SourceFileMap[field]
	if !ok {
		return ""
	} else {
		return data.getValue(field)
	}
}

func (sendData *SendData) parseField(s string) (f1, f2 string) {
	if s == "" {
		return
	}
	info := strings.Split(s, FILE_PARSE_SEP)
	if len(info) != 2 {
		return
	}
	return info[0], info[1]
}

func (sourceFile *SourceFile) getValue(field string) (value interface{}) {
	var (
		n = -1
	)
	for k, v := range sourceFile.Column {
		if v == field {
			n = k
		}
	}
	if n == -1 {
		return
	}
	sourceFile.m.Lock()
	if sourceFile.Index == uint64(len(sourceFile.Data)) {
		sourceFile.Index = 0
	}
	value = sourceFile.Data[sourceFile.Index][n]
	sourceFile.Index++
	sourceFile.m.Unlock()
	return
}
