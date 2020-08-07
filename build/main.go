package main

import (
	"gobom"
	"log"
)

func main() {
	if err := gobom.InitConfig("./config/app.toml"); err != nil {
		log.Fatal(err)
	}
	if err := gobom.InitDb("mysql", gobom.GetConfigs().MysqlSource); err != nil {
		log.Fatal(err)
	}
	gobom.GobomStore.AutoMigrate(map[string]gobom.TableAutoMigrateConfig{
		"script": {Model: &gobom.ScriptData{}},
		"task":   {Model: &gobom.TaskData{}},
	})
	api := gobom.NewApi()
	api.Http.Run(gobom.GetConfigs().ServerPort)
}
