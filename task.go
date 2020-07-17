package gobom

import (
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

var runTasks = make(map[string]*Task)
var mu sync.RWMutex

func GetRunTask(taskId string) (task *Task) {
	mu.RLock()
	defer mu.RUnlock()
	if task, ok := runTasks[taskId]; ok {
		return task
	} else {
		return nil
	}
}

func SetRunTask(task *Task) {
	mu.Lock()
	defer mu.Unlock()
	if _, ok := runTasks[task.TaskId]; !ok {
		runTasks[task.TaskId] = task
	}
}

func DelRunTask(taskId string) {
	mu.Lock()
	defer mu.Unlock()
	delete(runTasks, taskId)
}

type Task struct {
	TaskId string        `json:"taskId" gorm:"unique_index"`
	Worker *GobomRequest `json:"worker" gorm:"-"`
	Status int           `json:"status" gorm:"DEFAULT:0;"`
}

func NewTask(taskId string, opt *Options) (task *Task, err error) {
	gobomReq, err := NewGomBomRequest(opt)
	if err != nil {
		return nil, err
	}
	if taskId == "" {
		taskId = strconv.Itoa(int(utils.Now()))
	}
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
	if GetRunTask(task.TaskId) != nil {
		return ERR_TASK_RUN

	}
	task.SetStatus(STATUS_RUN)
	SetRunTask(task)
	task.Worker.Options.Init()
	return task.Worker.Dispose(func(err error) error {
		if err != nil {
			task.SetStatus(STATUS_ERROR)
		}
		if task.GetStatus() != STATUS_STOP {
			task.SetStatus(STATUS_OVER)
		}
		DelRunTask(task.TaskId)
		return err
	})
}

func (task *Task) Stop(count uint64) {
	if task == nil || task.Worker == nil {
		return
	}
	if count == CLOSE_ALL || count >= task.Worker.getConCurrent() {
		task.SetStatus(STATUS_STOP)
		DelRunTask(task.TaskId)
	}
	task.Worker.Close(count)
}

func (task *Task) Info() *Report {
	if task == nil || task.Worker == nil {
		return nil
	}
	return task.Worker.Info()
}

func (task *Task) GetStatus() int {
	return task.Status
}

func (task *Task) SetStatus(status int) {
	task.Status = status
}
