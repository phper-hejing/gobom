package gobom

import (
	"log"
	"reflect"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

var GobomStore, _ = InitSqlite("./gobom.db")

type Sqlite struct {
	db           *gorm.DB
	tableConfigs map[string]SqliteTableAutoMigrateConfig
}

type SqliteTableAutoMigrateConfig struct {
	Model interface{} // 表模型
}

func InitSqlite(source string) (sqlite *Sqlite, err error) {

	db, err := gorm.Open("sqlite3", source)
	if err != nil {
		return nil, err
	}
	if err := db.DB().Ping(); err != nil {
		log.Fatal(err)
	}
	return &Sqlite{
		db: db,
	}, nil
}

func (sqlite *Sqlite) GetDb() *gorm.DB {
	return sqlite.db
}

func (sqlite *Sqlite) Close() {
	sqlite.db.Close()
}

func (sqlite *Sqlite) GetTableName(model interface{}) string {
	for k, v := range sqlite.tableConfigs {
		if reflect.TypeOf(v.Model) == reflect.TypeOf(model) {
			return k
		}
	}
	return ""
}

func (sqlite *Sqlite) AutoMigrate(tableAutoMigrateConfig map[string]SqliteTableAutoMigrateConfig) error {
	for tablename, v := range tableAutoMigrateConfig {
		sqlite.db.Table(tablename).AutoMigrate(v.Model)
	}
	sqlite.tableConfigs = tableAutoMigrateConfig
	return nil
}
