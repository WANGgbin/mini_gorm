package tests

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"time"
)

var db *gorm.DB

func init() {
	var err error
	dsn := "test:123456@tcp(127.0.0.1:3306)/world?charset=utf8mb4&loc=Local&parseTime=true"
	db, err = gorm.Open(
		mysql.Open(dsn),
		&gorm.Config{
			NamingStrategy: schema.NamingStrategy{SingularTable: true},
			PrepareStmt: true,
		},
	)
	db = db.Debug()
	if err != nil {
		panic(fmt.Errorf("open mysql, error: %v", err))
	}
}

type person struct {
	ID uint64
	Name string
	Gender string `gorm:"default:female"` // 默认值: female
	Age uint16
	Secret []byte
	IsAlive bool
	BornTime time.Time
	UpdatedAt time.Time
	Extra *PersonExtra `gorm:"serializer:json"`
	DeletedAt gorm.DeletedAt
}

