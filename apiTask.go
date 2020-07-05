package gobom

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"io"
	"net/http"
)

type TaskData struct {
	gorm.Model
	Task    *Task  `json:"task" gorm:"EMBEDDED"`
	Options []byte `json:"options"`
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
		taskData.Task, err = NewTask(opt)
	} else {
		taskData.Task.TaskId = opt.TaskId
	}
	taskData.Options = opt.ToByte()

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
	if err != nil {
		return errors.New("创建任务实例失败")
	}
	go func() {
		taskData.Task.Run()
		taskData.Update()
	}()
	return
}

func (taskData *TaskData) Stop() (err error) {
	taskTemp, ok := runTasks.Load(taskData.Task.TaskId)
	if !ok {
		return errors.New("停止失败，任务没有运行")
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

func GetTask(data []byte) (task *Task, err error) {
	if data == nil {
		return nil, errors.New("opt is nil")
	}
	var opt Options
	if err := json.Unmarshal(data, &opt); err != nil {
		return nil, err
	}
	task, err = NewTask(&opt)
	if err != nil {
		return nil, err
	}
	return task, err
}
