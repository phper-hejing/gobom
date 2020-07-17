package gobom

import (
	"fmt"
	"testing"
	"time"
)

func TestTask(t *testing.T) {
	task, err := NewTask("", &Options{
		Url:        "https://www.jianshu.com/p/2360984a47a9",
		ConCurrent: 1,
		Duration:   5,
	})
	if err != nil {
		t.Error(err)
	}

	go func() {
		for {
			tt := GetRunTask(task.TaskId)
			fmt.Println(tt.Info())
			time.Sleep(time.Second)
		}
	}()

	task.Run()

}
