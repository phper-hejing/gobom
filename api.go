package gobom

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Api struct {
	Http *gin.Engine
}

type ApiReply struct {
	Code  int         `json:"code"`
	Msg   string      `json:"msg"`
	Data  interface{} `json:"data"`
	Error error       `json:"-"`
}

func NewApi() *Api {
	api := &Api{
		Http: gin.Default(),
	}
	api.Http.Use(func(ctx *gin.Context) {
		ctx.Header("Access-Control-Allow-Origin", "*")
		ctx.Header("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token, Authorization, Token")
		ctx.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, PATCH, DELETE")
		ctx.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		ctx.Header("Access-Control-Allow-Credentials", "true")

		if ctx.Request.Method == "OPTIONS" {
			ctx.AbortWithStatus(http.StatusNoContent)
		}

		// 处理请求
		ctx.Next()
	})
	api.RegisterRouter()
	return api
}

func (api *Api) RegisterRouter() {
	api.Http.Static("/frontend/", "./")
	api.Http.Any("/ws", TaskWsHandel)

	api.Http.Any("/task", TaskDataHandel)
	api.Http.Any("/task/add", TaskDataHandel)
	api.Http.Any("/task/edit", TaskDataHandel)
	api.Http.Any("/task/delete", TaskDataHandel)
	api.Http.Any("/task/run", TaskDataHandel)
	api.Http.Any("/task/info", TaskDataHandel)
	api.Http.Any("/task/stop", TaskDataHandel)

	api.Http.Any("/script", ScriptDataHandel)
	api.Http.Any("/script/add", ScriptDataHandel)
	api.Http.Any("/script/delete", ScriptDataHandel)
	api.Http.Any("/script/edit", ScriptDataHandel)
	api.Http.Any("/script/test", ScriptDataHandel)
}

func ApiResponse(ctx *gin.Context, reply ApiReply) {
	if reply.Error != nil {
		reply.Msg = reply.Error.Error()
	}
	ctx.JSON(http.StatusOK, reply)
}
