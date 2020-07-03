package gobom

import (
	"errors"
	"gobom/utils"
	"strconv"
	"sync"
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
	ERR_TASK_WORKER = errors.New("task worker is nil")
	ERR_TASK_RUN    = errors.New("任务正在运行中")
)

var runTasks sync.Map

type Task struct {
	TaskId string        `json:"taskId"`
	Worker *GobomRequest `json:"-"`
	Status int           `json:"status" gorm:"DEFAULT:0;"`
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

func (task *Task) Run() error {
	if task == nil || task.Worker == nil {
		return ERR_TASK_WORKER
	}
	if task.GetStatus() == STATUS_RUN {
		return ERR_TASK_RUN
	}
	task.SetStatus(STATUS_RUN)
	runTasks.Store(task.TaskId, task)
	task.Worker.Options.Init()
	return task.Worker.Dispose(func(err error) {
		if err != nil {
			task.SetStatus(STATUS_ERROR)
			return
		}
		if task.GetStatus() == STATUS_STOP {
			return
		}
		task.SetStatus(STATUS_OVER)
	})
}

func (task *Task) Stop(count uint64) {
	if task == nil || task.Worker == nil {
		return
	}
	if count == CLOSE_ALL || count >= task.Worker.getConCurrent() {
		task.SetStatus(STATUS_STOP)
		runTasks.Delete(task.TaskId)
	}
	task.Worker.Close(count)
}

func (task *Task) Info() string {
	if task == nil || task.Worker == nil {
		return ""
	}
	return task.Worker.Info()
}

func (task *Task) GetStatus() int {
	return task.Status
}

func (task *Task) SetStatus(status int) {
	task.Status = status
}
