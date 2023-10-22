package tests

import (
	"testing"
)

// gorm 允许通过 db.Set() 和 db.Get() 方法来传递值给钩子或者其他方法。
func TestSetOrGet(t *testing.T) {
	var p person
	if err := db.Set("gender", "male").Create(&p); err != nil {
		t.Fatalf("err: %v", err)
	}
}

//func (p *person) BeforeCreate(db *gorm.DB) error {
//	value, ok := db.Get("gender")
//	if !ok {
//		return errors.New("cant get value of gender")
//	}
//	if value != "female" {
//		return errors.New("only accept male")
//	}
//	return nil
//}
