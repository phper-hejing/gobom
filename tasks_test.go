package gobom

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestHttpWorker(t *testing.T) {

	var wg sync.WaitGroup
	var data = make([]*DataField, 3)
	data[0] = &DataField{
		Name: "test_a",
		Type: "int",
		Len:  65535,
	}
	data[1] = &DataField{
		Name:    "test_b",
		Type:    "string",
		Len:     10,
		Default: "123456",
		Dynamic: "",
	}
	data[2] = &DataField{
		Name:    "test_c",
		Type:    "file",
		Len:     10,
		Dynamic: "gm_admin.xlsx---username",
	}
	opt := &Options{
		Url:        "127.0.0.1:8888",
		ConCurrent: 1,
		Duration:   1,
		Interval:   0,
		Form:       FORM_TCP,
		SendData: &SendData{
			DataFieldList: data,
		},
	}
	opt.Init()
	task, err := NewTask(opt)
	if err != nil {
		t.Error(err)
		return
	}
	if err := task.Run(); err != nil {
		t.Error(err)
	}
	wg.Add(1)
	go func() {
		t := time.NewTicker(1 * time.Second)
		for {
			if task.GetStatus() == STATUS_NONE || task.GetStatus() == STATUS_OVER || task.GetStatus() == STATUS_ERROR {
				fmt.Println(task.Info())
				wg.Done()
				return
			}
			<-t.C
		}
	}()

	wg.Wait()
}
