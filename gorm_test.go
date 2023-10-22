package gorm

import (
	"testing"
	"time"
)

type person struct {
	ID       uint64 `gorm:"primaryKey;autoIncrement"`
	Name     string
	Gender   string	`gorm:"default:male"`
	Age      *uint16 `gorm:"default:18"`
	Secret   []byte
	IsAlive  bool
	BornTime time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `gorm:"softDelete"`
}

func TestGorm(t *testing.T) {

}
