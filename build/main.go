package main

import "gobom"

func main() {
	gobom.GobomStore.AutoMigrate(map[string]gobom.SqliteTableAutoMigrateConfig{
		"task":   {Model: &gobom.TaskData{}},
		"script": {Model: &gobom.ScriptData{}},
	})
	api := gobom.NewApi()
	api.Http.Run(":9600")
}
