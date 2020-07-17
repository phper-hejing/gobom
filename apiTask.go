package gobom

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/donnie4w/go-logger/logger"
	"github.com/gin-gonic/gin"
)

type TaskData struct {
	Model
	Name     string `json:"name"`
	Task     *Task  `json:"task" gorm:"EMBEDDED"`
	ScriptId uint   `json:"scriptId"`
	TaskJson string `json:"-"`
}

type TaskReqData struct {
	TaskId     string `json:"taskId"`
	Name       string `json:"name"`
	ConCurrent uint64 `json:"conCurrent"`
	Duration   uint64 `json:"duration"`
	ScriptId   uint   `json:"scriptId"`
}

func TaskDataHandel(ctx *gin.Context) {
	var reqParam TaskReqData
	opt := &Options{}
	script := &ScriptData{}
	taskData := TaskData{
		Task: &Task{},
	}
	var msg string
	var err error
	var data interface{}
	defer func() {
		if err != nil {
			msg = err.Error()
		}
		ctx.JSON(http.StatusOK, &ApiReply{
			Msg:  msg,
			Data: data,
		})
	}()
	if err = ctx.ShouldBind(&reqParam); err != nil {
		if err != io.EOF {
			return
		}
	}

	taskData.Task.TaskId = reqParam.TaskId
	if reqParam.ScriptId != 0 {
		taskData.First()
		if reqParam.ScriptId != taskData.ScriptId {
			script.ID = reqParam.ScriptId
			if script, err = script.First(); err != nil {
				return
			}
			if err = json.Unmarshal([]byte(script.Data), opt); err != nil {
				return
			}
			taskData.Name = reqParam.Name
			taskData.ScriptId = reqParam.ScriptId
			opt.ConCurrent = reqParam.ConCurrent
			opt.Duration = reqParam.Duration
			if taskData.Task, err = NewTask(taskData.Task.TaskId, opt); err != nil {
				return
			}
		}
	}

	switch ctx.FullPath() {
	case "/task":
		if taskData.Task.TaskId == "" {
			data, err = taskData.Get()
		} else {
			data, err = taskData.First()
		}
		if err != nil {
			data = nil
		}
	case "/task/add":
		err = taskData.Add()
	case "/task/edit":
		err = taskData.Edit(reqParam)
	case "/task/delete":
		if task := GetRunTask(taskData.Task.TaskId); task != nil {
			err = errors.New("请先停止任务")
			return
		}
		err = taskData.Del()
	case "/task/run":
		err = taskData.Run()
	case "/task/info":
		data, err = taskData.Info()
	case "/task/stop":
		err = taskData.Stop()
	}
}

func (taskData *TaskData) BeforeCreate() (err error) {
	bt, err := json.Marshal(taskData.Task)
	taskData.TaskJson = string(bt)
	return err
}

func (taskData *TaskData) BeforeSave() (err error) {
	bt, err := json.Marshal(taskData.Task)
	taskData.TaskJson = string(bt)
	return err
}

func (taskData *TaskData) AfterFind() (err error) {
	defer func() {
		if task := GetRunTask(taskData.Task.TaskId); task != nil && err == nil {
			taskData.Task.Status = STATUS_RUN
		}
	}()
	return json.Unmarshal([]byte(taskData.TaskJson), taskData.Task)
}

func (taskData *TaskData) Add() (err error) {
	return GobomStore.GetDb().Table(GobomStore.GetTableName(taskTable)).Create(taskData).Error
}

func (taskData *TaskData) Edit(reqParam TaskReqData) (err error) {
	var t TaskData
	if err := GobomStore.GetDb().Table(GobomStore.GetTableName(taskTable)).Where("task_id = ?", taskData.Task.TaskId).First(&t).Error; err != nil {
		return err
	}
	t.Name = reqParam.Name
	t.ScriptId = reqParam.ScriptId
	t.Task = taskData.Task
	t.Task.Worker.setConCurrent(reqParam.ConCurrent)
	t.Task.Worker.setDuration(reqParam.Duration)
	return GobomStore.GetDb().Table(GobomStore.GetTableName(taskTable)).Where("task_id = ?", taskData.Task.TaskId).Save(&t).Error
}

func (taskData *TaskData) Del() (err error) {
	return GobomStore.GetDb().Table(GobomStore.GetTableName(taskTable)).Where("task_id = ?", taskData.Task.TaskId).Delete(taskData).Error
}

func (taskData *TaskData) Update() (err error) {
	return GobomStore.GetDb().Table(GobomStore.GetTableName(taskTable)).Where("task_id = ?", taskData.Task.TaskId).Save(taskData).Error
}

func (taskData *TaskData) First() (taskDataList *TaskData, err error) {
	return taskData, GobomStore.GetDb().Table(GobomStore.GetTableName(taskTable)).Where("task_id = ?", taskData.Task.TaskId).First(taskData).Error
}

func (taskData *TaskData) Get() (taskDataList []TaskData, err error) {
	var taskDataListTemp []TaskData
	err = GobomStore.GetDb().Table(GobomStore.GetTableName(taskTable)).Find(&taskDataListTemp).Error
	return taskDataListTemp, err
}

func (taskData *TaskData) Run() (err error) {
	taskData.First()
	go func() {
		if err := taskData.Task.Run(); err != nil {
			logger.Debug(err)
		}
		taskData.Update()
	}()
	return
}

func (taskData *TaskData) Stop() (err error) {
	task := GetRunTask(taskData.Task.TaskId)
	task.Stop(CLOSE_ALL)
	return
}

func (taskData *TaskData) Info() (data interface{}, err error) {
	task := GetRunTask(taskData.Task.TaskId)
	if task != nil {
		taskData.Task = task
	} else {
		taskData, err = taskData.First()
		if err != nil {
			return nil, err
		}
	}
	return taskData, nil
}
