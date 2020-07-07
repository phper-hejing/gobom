package gobom

import (
	"encoding/json"
	"github.com/donnie4w/go-logger/logger"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"io"
	"net/http"
)

type TaskData struct {
	gorm.Model
	Task     *Task  `json:"task" gorm:"EMBEDDED"`
	TaskByte []byte `json:"taskByte"`
}

var taskTable = &TaskData{}

func TaskDataHandel(ctx *gin.Context) {
	opt := &Options{}
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
	if err = ctx.ShouldBind(opt); err != nil {
		if err != io.EOF {
			return
		}
	}

	if opt.TaskId == "" {
		if taskData.Task, err = NewTask(opt); err != nil {
			return
		}
	} else {
		taskData.Task.TaskId = opt.TaskId
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
	case "/task/del":
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
	taskData.TaskByte, err = json.Marshal(taskData.Task)
	return err
}

func (taskData *TaskData) BeforeSave() (err error) {
	taskData.TaskByte, err = json.Marshal(taskData.Task)
	return err
}

func (taskData *TaskData) AfterFind() (err error) {
	return json.Unmarshal(taskData.TaskByte, taskData.Task)
}

func (taskData *TaskData) Add() (err error) {
	return GobomStore.GetDb().Table(GobomStore.GetTableName(taskTable)).Create(taskData).Error
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
	taskTemp, ok := runTasks.Load(taskData.Task.TaskId)
	if !ok {
		return ERR_TASK_STOP_NONE
	}
	task := taskTemp.(*Task)
	task.Stop(CLOSE_ALL)
	return
}

func (taskData *TaskData) Info() (data string, err error) {
	taskTemp, ok := runTasks.Load(taskData.Task.TaskId)
	if ok {
		task := taskTemp.(*Task)
		return task.Info(), nil
	}
	taskData, err = taskData.First()
	if err != nil {
		return "", err
	}
	return taskData.Task.Info(), nil
}
