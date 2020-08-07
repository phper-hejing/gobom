package gobom

import (
	"database/sql/driver"
	"fmt"
	"github.com/donnie4w/go-logger/logger"
	"reflect"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

type TableAutoMigrateConfig struct {
	Model interface{} // 表模型
}

type DbInterface interface {
	GetDb() *gorm.DB
	Close()
	GetTableName(model interface{}) string
	AutoMigrate(tableAutoMigrateConfig map[string]TableAutoMigrateConfig)
}

var GobomStore *GobomDb

type GobomDb struct {
	Db           *gorm.DB
	TableConfigs map[string]TableAutoMigrateConfig
}

func InitDb(kind, source string) error {
	db, err := gorm.Open(kind, source)
	if err != nil {
		return err
	}
	if err := db.DB().Ping(); err != nil {
		return err
	}
	db.DB().SetMaxOpenConns(20)
	db.DB().SetMaxIdleConns(10)
	GobomStore = &GobomDb{
		Db: db,
	}
	return nil
}

func (gobomDb *GobomDb) GetDb() *gorm.DB {
	return gobomDb.Db
}

func (gobomDb *GobomDb) Close() {
	gobomDb.Db.Close()
}

func (gobomDb *GobomDb) GetTableName(model interface{}) string {
	for k, v := range gobomDb.TableConfigs {
		if reflect.TypeOf(v.Model) == reflect.TypeOf(model) {
			return k
		}
	}
	return ""
}

func (gobomDb *GobomDb) AutoMigrate(tableAutoMigrateConfig map[string]TableAutoMigrateConfig) {
	for tablename, v := range tableAutoMigrateConfig {
		if err := gobomDb.Db.Set("gorm:table_options", "ENGINE=InnoDB").Table(tablename).AutoMigrate(v.Model).Error; err != nil {
			logger.Fatal(err)
		}
	}
	gobomDb.TableConfigs = tableAutoMigrateConfig
}

type Model struct {
	ID        uint `gorm:"primary_key"`
	CreatedAt JSONTime
	UpdatedAt JSONTime
	DeletedAt *JSONTime `sql:"index"`
}

// JSONTime format json time field by myself
type JSONTime struct {
	time.Time
}

// MarshalJSON on JSONTime format Time field with %Y-%m-%d %H:%M:%S
func (t JSONTime) MarshalJSON() ([]byte, error) {
	if (t == JSONTime{}) {
		formatted := fmt.Sprintf("\"%s\"", "")
		return []byte(formatted), nil
	} else {
		formatted := fmt.Sprintf("\"%s\"", t.Format("2006-01-02 15:04:05"))
		return []byte(formatted), nil
	}
}

// Value insert timestamp into mysql need this function.
func (t JSONTime) Value() (driver.Value, error) {
	var zeroTime time.Time
	if t.Time.UnixNano() == zeroTime.UnixNano() {
		return nil, nil
	}
	return t.Time, nil
}

// Scan valueof time.Time
func (t *JSONTime) Scan(v interface{}) error {
	value, ok := v.(time.Time)
	if ok {
		*t = JSONTime{Time: value}
		return nil
	}
	return fmt.Errorf("can not convert %v to timestamp", v)
}
