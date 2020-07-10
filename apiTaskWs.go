package gobom

import (
	"encoding/json"
	"github.com/donnie4w/go-logger/logger"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gobom/utils"
	"net/http"
	"sync"
	"time"
)

const (
	TYPE_CLIENT = iota + 1
	TYPE_SERVER
)

const (
	WS_PING = iota + 1
	WS_TASK_RUN
	WS_TASK_STOP
	WS_TASK_REPORT
)

type TaskWs struct {
	Conn       *websocket.Conn `json:"-"`
	mu         sync.Mutex
	TaskWsData *TaskWsData `json:"taskWsData"`
}

type TaskWsData struct {
	Type  int         `json:"type"`
	Data  interface{} `json:"data"`
	Error string      `json:"error"`
}

var taskTable = &TaskData{}
var upGrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func TaskWsHandel(ctx *gin.Context) {
	ws, err := upGrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		logger.Debug(err)
		return
	}
	defer ws.Close()
	taskWs := &TaskWs{
		TaskWsData: &TaskWsData{},
		mu:         sync.Mutex{},
	}
	taskWs.Conn = ws
	go taskWs.Ping()
	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			logger.Debug(err)
			return
		}
		if err := json.Unmarshal(msg, taskWs.TaskWsData); err != nil {
			logger.Debug(err)
		}
		taskWs.ParseMsg()
	}
}

func (taskWs *TaskWs) ParseMsg() {
	var err error
	var data interface{}

	msgData, _ := taskWs.TaskWsData.Data.(map[string]interface{})
	taskId, _ := msgData["taskId"].(string)
	taskData := &TaskData{
		Task: &Task{
			TaskId: taskId,
		},
	}
	switch taskWs.TaskWsData.Type {
	case WS_TASK_RUN:
		err = taskData.Run()
		data = map[string]string{"taskId": taskId}
	case WS_TASK_STOP:
		err = taskData.Stop()
		data = map[string]string{"taskId": taskId}
	case WS_TASK_REPORT:
		data, err = taskData.Info()
	default:
		return
	}

	taskWs.TaskWsData = &TaskWsData{
		Type:  taskWs.TaskWsData.Type,
		Data:  data,
		Error: utils.GetErrString(err),
	}
	taskWs.SendMsg()
}

func (taskWs *TaskWs) SendMsg() error {
	bt, err := json.Marshal(taskWs.TaskWsData)
	if err != nil {
		logger.Debug(err)
		return err
	}
	taskWs.mu.Lock()
	defer taskWs.mu.Unlock()
	if err := taskWs.Conn.WriteMessage(websocket.TextMessage, bt); err != nil {
		logger.Debug(err)
		return err
	}
	return nil
}

func (taskWs *TaskWs) Ping() {
	for {
		if taskWs.Conn == nil {
			return
		}
		time.Sleep(time.Duration(1) * time.Second)
		taskWs.TaskWsData = &TaskWsData{
			Type: WS_PING,
		}
		if err := taskWs.SendMsg(); err != nil {
			return
		}
	}
}
