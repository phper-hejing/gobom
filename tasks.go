package gobom

import "sync"

const (
	STATUS_NONE = iota
	STATUS_RUN
	STATUS_OVER
)

type TaskList struct {
	Tasks map[string]*Task
	cond  *sync.Cond
	mu    *sync.Mutex
}

type Task struct {
	id     string
	worker *GobomRequest
	status int
}

func (taskList *TaskList) waitForRun() {
	taskList.cond.L.Lock()
	defer taskList.cond.L.Unlock()
	taskList.cond.Wait()
}

func (taskList *TaskList) Add(task *Task) {
	taskList.mu.Lock()
	defer taskList.mu.Unlock()
	taskList.Tasks[task.id] = task
	taskList.cond.Signal()
}

func (taskList *TaskList) Del(id string) {
	taskList.mu.Lock()
	defer taskList.mu.Unlock()
	if taskList.Tasks[id].status == STATUS_RUN {
		taskList.Tasks[id].stop()
	}
	delete(taskList.Tasks, id)
}

func (taskList *TaskList) Run(id string) {
	taskList.mu.Lock()
	tasks := taskList.Tasks
	taskList.mu.Unlock()

	taskList.cond.L.Lock()
	taskList.cond.Wait()
	for _, task := range tasks {
		task.exec()
		taskList.Del(task.id)
	}
	taskList.cond.L.Unlock()
}

func (task *Task) exec() {
	if task.status == STATUS_OVER {
		return
	}
	task.status = STATUS_RUN
	task.worker.Dispose(task.id)
	task.status = STATUS_OVER
}

func (task *Task) stop() {
	task.worker.Close()
	task.status = STATUS_OVER
}
