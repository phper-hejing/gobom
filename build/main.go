package main

import "gobom"

func main() {
	gobom.GobomStore.AutoMigrate(map[string]gobom.SqliteTableAutoMigrateConfig{
		"task":   {Model: &gobom.Task{}},
		"script": {Model: &gobom.Script{}},
	})
	api := gobom.NewApi()
	api.Http.Run(":9600")
}
