package gobom

import (
	"encoding/json"
	"errors"
	"github.com/donnie4w/go-logger/logger"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
)

type TaskData struct {
	Model
	Name     string `json:"name" gorm:"unique_index"`
	Task     *Task  `json:"task" gorm:"EMBEDDED"`
	ScriptId uint   `json:"scriptId"`
	TaskJson string `json:"-" gorm:"type:longtext"`
}

type TaskReqData struct {
	TaskId     string `json:"taskId"`
	Name       string `json:"name"`
	ConCurrent uint64 `json:"conCurrent"`
	Duration   uint64 `json:"duration"`
	ScriptId   uint   `json:"scriptId"`
}

var taskTable = &TaskData{}

func TaskDataHandel(ctx *gin.Context) {
	var reqParam TaskReqData
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

	taskData.Name = reqParam.Name
	taskData.Task.TaskId = reqParam.TaskId

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
		var task *Task
		task, err = taskData.InitTask(reqParam)
		if err != nil {
			return
		}
		taskData.Task = task
		taskData.ScriptId = reqParam.ScriptId
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
		if task := GetRunTask(taskData.Task.TaskId); task != nil {
			err = errors.New("任务正在运行")
			return
		}
		err = taskData.Run()
	case "/task/info":
		//data, err = taskData.Info()
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
	t.Task.Worker.setConCurrent(reqParam.ConCurrent)
	t.Task.Worker.setDuration(reqParam.Duration)
	if t.ScriptId != reqParam.ScriptId { // 修改了脚本ID需要重新初始化任务实例
		task, err := t.InitTask(reqParam)
		if err != nil {
			return err
		}
		t.Task = task
		t.ScriptId = reqParam.ScriptId
	}
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
	if task == nil {
		return errors.New("任务没有运行")
	}
	task.Stop(CLOSE_ALL)
	return
}

func (taskData *TaskData) Info() (data interface{}, err error) {
	var task *Task
	task = GetRunTask(taskData.Task.TaskId)

	if task == nil {
		taskData, err = taskData.First()
		if err != nil || taskData.Task == nil {
			logger.Debug("taskData.Task is nil", err)
			return
		}
		task = taskData.Task
	}

	if task.Worker == nil || task.Worker.Report == nil {
		logger.Debug("report is nil")
		return
	}

	return &TaskData{
		Name: taskData.Name,
		Task: &Task{
			TaskId: task.TaskId,
			Worker: &GobomRequest{
				Report:     task.Worker.Report.Copy(),
				Duration:   task.Worker.Duration,
				ConCurrent: task.Worker.ConCurrent,
			},
			Status: task.Status,
		},
	}, nil
}

func (taskData *TaskData) InitTask(reqParam TaskReqData) (task *Task, err error) {
	var script = &ScriptData{}
	var opt = &Options{}
	script.ID = reqParam.ScriptId
	if script, err = script.First(); err != nil {
		return
	}

	if err = json.Unmarshal([]byte(script.Data), opt); err != nil {
		return
	}

	opt.ConCurrent = reqParam.ConCurrent
	opt.Duration = reqParam.Duration
	return NewTask(taskData.Task.TaskId, opt)
}

func ResetTaskScript(scriptId uint) {
	var taskDataList []TaskData
	if err := GobomStore.GetDb().Table(GobomStore.GetTableName(taskTable)).Where("script_id = ?", scriptId).Find(&taskDataList).Error; err == nil {
		for _, data := range taskDataList {
			task, err := data.InitTask(TaskReqData{
				ConCurrent: data.Task.Worker.getConCurrent(),
				Duration:   data.Task.Worker.getDuration(),
				ScriptId:   scriptId,
			})
			if err != nil {
				logger.Debug(err)
				return
			}
			data.Task = task
			data.ScriptId = scriptId
			GobomStore.GetDb().Table(GobomStore.GetTableName(taskTable)).Save(&data)
		}
	}
}
