package gobom

import (
	"errors"
	"gobom/utils"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Api struct {
	Http *gin.Engine
}

type IMessage interface {
	Init(*gin.Context) error
	Do()
}

type ApiMessage struct {
	ctx *gin.Context
	opt *Options
}

type ApiReply struct {
	Code  int         `json:"code"`
	Msg   string      `json:"msg"`
	Data  interface{} `json:"data"`
	Error error       `json:"-"`
}

type ApiMessageTask struct {
	ApiMessage
}

func NewApi() *Api {
	api := &Api{
		Http: gin.Default(),
	}
	api.Http.Use(Cors())
	api.RegisterRouter()
	return api
}

func Cors() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Header("Access-Control-Allow-Origin", "*")
		ctx.Header("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token, Authorization, Token")
		ctx.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, PATCH, DELETE")
		ctx.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		ctx.Header("Access-Control-Allow-Credentials", "true")

		// 放行所有OPTIONS方法，因为有的模板是要请求两次的
		if ctx.Request.Method == "OPTIONS" {
			ctx.AbortWithStatus(http.StatusNoContent)
		}

		// 处理请求
		ctx.Next()
	}
}

func (apiMessage *ApiMessage) Init(context *gin.Context) (err error) {
	opt := Options{}
	if err := context.ShouldBind(&opt); err != nil {
		if err != io.EOF {
			return utils.Err(err)
		}
	}
	apiMessage.ctx = context
	apiMessage.opt = &opt
	return nil
}

func HandelMessage(apiMsg IMessage) gin.HandlerFunc {
	return func(context *gin.Context) {
		if err := apiMsg.Init(context); err != nil {
			context.JSON(http.StatusOK, ApiReply{
				Msg: "参数解析错误",
			})
			utils.ErrPrint(err)
		}
		apiMsg.Do()
	}
}

func ApiResponse(ctx *gin.Context, reply ApiReply) {
	if reply.Error != nil {
		reply.Msg = reply.Error.Error()
	}
	ctx.JSON(http.StatusOK, reply)
}

func (api *Api) RegisterRouter() {
	api.Http.Any("task", HandelMessage(new(ApiMessageTask)))
	api.Http.Any("task/add", HandelMessage(new(ApiMessageTask)))
	api.Http.Any("task/del", HandelMessage(new(ApiMessageTask)))
	api.Http.Any("task/run", HandelMessage(new(ApiMessageTask)))
	api.Http.Any("task/info", HandelMessage(new(ApiMessageTask)))
	api.Http.Any("task/stop", HandelMessage(new(ApiMessageTask)))

	api.Http.Any("/script", ScriptApi)
	api.Http.Any("/script/add", ScriptApi)
	api.Http.Any("/script/delete", ScriptApi)
	api.Http.Any("/script/edit", ScriptApi)
}

func (apiMsg *ApiMessageTask) Do() {
	var err error
	var msg string
	var data interface{}
	var task *Task
	defer func() {
		if err != nil {
			msg = err.Error()
		}
		ApiResponse(apiMsg.ctx, ApiReply{
			Msg:  msg,
			Data: data,
		})
	}()
	switch apiMsg.ctx.FullPath() {
	case "/task":
		if apiMsg.opt.TaskId == "" {
			data, err = TaskFindAll()
		} else {
			data, err = TaskFind(apiMsg.opt.TaskId)
		}
	case "/task/add":
		task, err = NewTask(apiMsg.opt)
		if err != nil {
			return
		}
		err = TaskAdd(task)
		data = map[string]string{"taskId": task.TaskId}
	case "/task/del":
		if apiMsg.opt.TaskId == "" {
			err = errors.New("任务id为空")
			return
		}
		err = TaskDel(apiMsg.opt.TaskId)
	case "/task/run":
		if apiMsg.opt.TaskId != "" {
			if GetRunTaskId() == apiMsg.opt.TaskId { // 运行中的任务增加或减少压测协程数
				if apiMsg.opt.LessenConCurrent > 0 {
					runTask.Stop(apiMsg.opt.LessenConCurrent)
				} else if apiMsg.opt.ConCurrent > 0 {
					runTask.Worker.AddConcurrentAndStart(apiMsg.opt.ConCurrent)
				} else {
					err = ERR_TASK_STOP_RUN
					return
				}
			} else { // 中止的任务继续执行
				task, err = TaskFind(apiMsg.opt.TaskId)
				if err != nil {
					return
				}
			}
		} else {
			task, err = NewTask(apiMsg.opt)
			if err != nil {
				return
			}
			TaskAdd(task)
		}

		if err = task.Run(); err != nil {
			return
		}
		data = map[string]string{"taskId": task.TaskId}
	case "/task/info":
		if apiMsg.opt.TaskId == "" {
			err = errors.New("任务id为空")
			return
		}
		task, err := TaskFind(apiMsg.opt.TaskId)
		if err != nil {
			return
		}
		data, err = task.Info()
	case "/task/stop":
		TaskStop(apiMsg.opt.TaskId)
	}

}
