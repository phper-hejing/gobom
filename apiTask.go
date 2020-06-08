package gobom

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/donnie4w/go-logger/logger"
	"github.com/jinzhu/gorm"
	"strconv"

	"gobom/utils"
)

const (
	STATUS_NONE = iota
	STATUS_WAIT
	STATUS_RUN
	STATUS_OVER
	STATUS_STOP
	STATUS_ERROR
)

var (
	ERR_TASK_STOP_RUN = errors.New("请先关闭正在运行的任务")
	ERR_TASK_RUN      = errors.New("任务正在运行")
	ERR_TASK_OVER     = errors.New("任务不存在")
)

var taskTable = &Task{}
var runTask *Task

type Task struct {
	gorm.Model
	TaskId string `gorm:"NOT NULL;"`
	Worker *GobomRequest
	Status int `gorm:"DEFAULT:0;"`

	WorkerByte []byte `json:"workerJson" gorm:"type:'blob'"`
}

func (task *Task) BeforeCreate(scope *gorm.Scope) error {
	data, err := json.Marshal(task.Worker)
	scope.SetColumn("WorkerByte", data)
	return err
}

func (task *Task) BeforeUpdate(scope *gorm.Scope) error {
	data, err := json.Marshal(task.Worker)
	scope.SetColumn("WorkerByte", data)
	return err
}

func (task *Task) AfterFind(scope *gorm.Scope) error {
	task.Worker = &GobomRequest{}
	err := json.Unmarshal(task.WorkerByte, task.Worker)
	task.WorkerByte = nil
	return err
}

func NewTask(opt *Options) (task *Task, err error) {
	gobomReq, err := NewGomBomRequest(opt)
	if err != nil {
		return nil, err
	}
	taskId := strconv.Itoa(int(utils.Now()))
	return &Task{
		TaskId: taskId,
		Worker: gobomReq,
		Status: STATUS_WAIT,
	}, nil
}

func TaskAdd(task *Task) error {
	if err := GobomStore.GetDb().Table(GobomStore.GetTableName(taskTable)).Create(task).Error; err != nil {
		return err
	}
	return nil
}

func TaskStop(taskId string) error {
	if GetRunTaskId() == taskId {
		runTask.Stop(CLOSE_ALL)
	}
	return nil
}

func TaskDel(taskId string) error {
	var task Task
	if GobomStore.GetDb().Table(GobomStore.GetTableName(taskTable)).Where("task_id = ?", taskId).First(&task).RowsAffected == 0 {
		return errors.New("任务不存在")
	}
	if err := GobomStore.GetDb().Table(GobomStore.GetTableName(taskTable)).Delete(&task).Error; err != nil {
		return err
	}
	if task.TaskId == GetRunTaskId() {
		runTask.Stop(CLOSE_ALL)
	}
	return nil
}

func TaskFind(taskId string) (task *Task, err error) {
	task = &Task{}
	if err := GobomStore.GetDb().Table(GobomStore.GetTableName(taskTable)).Where("task_id = ?", taskId).First(task).Error; err != nil {
		return nil, err
	}
	return
}

func TaskFindAll() (task []Task, err error) {
	if err := GobomStore.GetDb().Table(GobomStore.GetTableName(taskTable)).First(&task).Error; err != nil {
		return nil, err
	}
	return
}

func TaskAddConcurrent(taskId string, count uint64) (err error) {
	if GetRunTaskId() == taskId {
		runTask.Worker.Start(count)
	}
	return nil
}

func (task *Task) Run() error {

	if task == nil || task.Worker == nil {
		return ERR_TASK_OVER
	}

	if GetRunTaskId() != "" {
		return ERR_TASK_STOP_RUN
	}

	task.Worker.Options.Init()
	go func() {
		status := STATUS_RUN
		runTask = task
		task.SetStatus(status)                        // 设置任务为运行状态
		if err := task.Worker.Dispose(); err != nil { // 执行任务,挂起
			status = STATUS_ERROR
			logger.Debug(fmt.Sprintf("任务id：%s, 异常结束：%s", task.TaskId, err.Error()))
		} else {
			if task.GetStatus() == STATUS_RUN {
				status = STATUS_OVER
				logger.Debug(fmt.Sprintf("任务id：%s, 正常结束", task.TaskId))
			} else if task.GetStatus() == STATUS_STOP {
				status = STATUS_STOP
				logger.Debug(fmt.Sprintf("任务id：%s, 手动结束", task.TaskId))
			}
		}
		runTask = nil
		task.SetStatus(status)
	}()
	return nil
}

func (task *Task) Stop(count uint64) {
	if task == nil || task.Worker == nil {
		return
	}
	if count == CLOSE_ALL || count >= task.Worker.getConCurrent() {
		task.SetStatus(STATUS_STOP)
	}
	task.Worker.Close(count)
}

func (task *Task) Info() (string, error) {
	if task == nil || task.Worker == nil {
		return "{}", ERR_TASK_OVER
	}
	if task.TaskId == GetRunTaskId() {
		return runTask.Worker.Info(), nil
	}
	return task.Worker.Info(), nil
}

func (task *Task) GetStatus() int {
	if task == nil {
		return STATUS_NONE
	}
	return task.Status
}

func (task *Task) SetStatus(status int) {
	if task == nil {
		return
	}
	if err := GobomStore.GetDb().Table(GobomStore.GetTableName(taskTable)).Model(task).Where("task_id = ?", task.TaskId).Updates(&Task{
		Status:     status,
		WorkerByte: task.WorkerByte,
	}).Error; err != nil {
		logger.Debug(err)
	}
	task.Status = status
}

func GetRunTaskId() string {
	if runTask == nil {
		return ""
	}
	return runTask.TaskId
}
