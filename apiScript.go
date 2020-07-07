package gobom

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"io"
	"net/http"
)

type ScriptData struct {
	gorm.Model
	Type     string `json:"type"`
	Name     string `json:"name"`
	Protocol int    `json:"protocol"`
	Data     string `json:"data" gorm:"type:'longtext'"`
}

var scriptTable = &ScriptData{}

func ScriptDataHandel(ctx *gin.Context) {
	scriptData := ScriptData{}
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
	if err = ctx.ShouldBind(&scriptData); err != nil {
		if err != io.EOF {
			return
		}
	}
	switch ctx.FullPath() {
	case "/script":
		if scriptData.ID == 0 {
			data, err = scriptData.Get()
		} else {
			data, err = scriptData.First()
		}
	case "/script/add":
		err = scriptData.Add()
	case "/script/delete":
		err = scriptData.Del()
	case "/script/edit":
		err = scriptData.Update()
	case "/script/test":
		err = scriptData.Run()
	}

}

func (scriptData *ScriptData) Run() error {
	var opt Options
	if scriptData.Data == "" {
		return ERR_PARAM
	}
	if err := json.Unmarshal([]byte(scriptData.Data), &opt); err != nil {
		return ERR_PARAM_PARSE
	}
	opt.Form = scriptData.Protocol
	gobomReq, err := NewGomBomRequest(&opt)
	if err != nil {
		return err
	}
	return gobomReq.boardTest()
}

func (scriptData *ScriptData) Add() error {
	if err := GobomStore.GetDb().Table(GobomStore.GetTableName(scriptTable)).Create(scriptData).Error; err != nil {
		return err
	}
	return nil
}

func (scriptData *ScriptData) Del() error {
	if err := GobomStore.GetDb().Table(GobomStore.GetTableName(scriptTable)).Delete(&scriptData).Error; err != nil {
		return err
	}
	return nil
}

func (scriptData *ScriptData) Update() error {
	if err := GobomStore.GetDb().Table(GobomStore.GetTableName(scriptTable)).Save(scriptData).Error; err != nil {
		return err
	}
	return nil
}

func (scriptData *ScriptData) First() (*ScriptData, error) {
	if err := GobomStore.GetDb().Table(GobomStore.GetTableName(scriptTable)).First(scriptData).Error; err != nil {
		return nil, err
	}
	return scriptData, nil
}

func (scriptData *ScriptData) Get() ([]ScriptData, error) {
	var scripts []ScriptData
	if err := GobomStore.GetDb().Table(GobomStore.GetTableName(scriptTable)).Find(&scripts).Error; err != nil {
		return nil, err
	}
	return scripts, nil
}
