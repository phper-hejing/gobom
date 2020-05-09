package gobom

const (
	FORM_HTTP = iota
	FORM_TCP
	FORM_WEBSOCKET
)

type Options struct {
	ConCurrency uint64 `json:"conCurrent"` // 并发数
	Duration    uint64 `json:"duration"`   // 持续时间（秒）
	Interval    uint64 `json:"interval"`   // 请求间隔时间
	Form        int    `json:"form"`       // http|websocket|tcp
}
