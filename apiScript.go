package gobom

import (
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"io"
	"net/http"
)

type Script struct {
	gorm.Model
	Type     string `json:"type"`
	Name     string `json:"name"`
	Protocol int    `json:"protocol"`
	Data     string `json:"data" gorm:"type:'longtext'"`
}

var scriptTable = &Script{}

func ScriptApi(ctx *gin.Context) {
	script := Script{}
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
	if err = ctx.ShouldBind(&script); err != nil {
		if err != io.EOF {
			return
		}
	}
	switch ctx.FullPath() {
	case "/script":
		if script.ID != 0 {
			data, err = script.ScriptFind()
		} else {
			data, err = script.ScriptFindAll()
		}
	case "/script/add":
		err = script.ScriptAdd()
	case "/script/delete":
		err = script.ScriptDel()
	case "/script/edit":
		err = script.ScriptEdit()
	}

}

func (script *Script) ScriptAdd() error {
	if err := GobomStore.GetDb().Table(GobomStore.GetTableName(scriptTable)).Create(script).Error; err != nil {
		return err
	}
	return nil
}

func (script *Script) ScriptDel() error {
	if err := GobomStore.GetDb().Table(GobomStore.GetTableName(scriptTable)).Delete(&script).Error; err != nil {
		return err
	}
	return nil
}

func (script *Script) ScriptEdit() error {
	if err := GobomStore.GetDb().Table(GobomStore.GetTableName(scriptTable)).Save(script).Error; err != nil {
		return err
	}
	return nil
}

func (script *Script) ScriptFind() (*Script, error) {
	if err := GobomStore.GetDb().Table(GobomStore.GetTableName(scriptTable)).First(script).Error; err != nil {
		return nil, err
	}
	return script, nil
}

func (script *Script) ScriptFindAll() ([]Script, error) {
	var scripts []Script
	if err := GobomStore.GetDb().Table(GobomStore.GetTableName(scriptTable)).Find(&scripts).Error; err != nil {
		return nil, err
	}
	return scripts, nil
}
