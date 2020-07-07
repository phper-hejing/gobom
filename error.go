package gobom

import "errors"

var (
	ERR_NONE = false

	ERR_TASK_WORKER = errors.New("task worker is nil")
	ERR_TASK_RUN    = errors.New("任务正在运行中")

	ERR_PARAM       = errors.New("参数错误")
	ERR_PARAM_PARSE = errors.New("参数解析错误")

	ERR_URL         = errors.New("URL不能为空")
	ERR_CONCURRENCY = errors.New("并发数数值过小")
	ERR_DURATION    = errors.New("持续时间数值过小")
	ERR_FORM        = errors.New("无法识别的请求类型")
	ERR_CONCURRENT  = errors.New("并发数不能为0")
	ERR_OPTIONS_NIL = errors.New("options is nil")

	ERR_FILE_INIT  = errors.New("初始化失败")
	ERR_FILE_PARSE = errors.New("解析文件数据失败")
	ERR_FILE_OPEN  = errors.New("打开文件数据失败")
	ERR_FILE_READ  = errors.New("读取文件数据失败")

	ERR_TASK_CREATE    = errors.New("创建任务实例失败")
	ERR_TASK_STOP_NONE = errors.New("停止失败，任务没有运行")
)
