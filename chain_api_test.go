package gorm

import (
	"github.com/WANGgbin/mini_gorm/clause"
	"github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestDB_Where(t *testing.T) {
	convey.Convey("", t, func() {
		db := &DB{}
		db.stmt = newStmt(db)

		tx := db.Where(
			db.Where("pizza = ?", "pepp").Where(
				db.Where("size = ?", "small").Or("size = ?", "medium"),
			),
		).Or(
			db.Where("pizza = ?", "hawai").Where("size = ?", "xlarge"),
		)
		tx.stmt.setWhereClause()

		if tx.err != nil {
			t.Fatal(tx.err)
		}

		convey.So(tx.stmt.css[clause.KindWhere].GetContentWithPlaceHolder(), convey.ShouldEqual,
			`((pizza = ?) AND ((size = ?) OR (size = ?))) OR ((pizza = ?) AND (size = ?))`)
		convey.So(tx.stmt.css[clause.KindWhere].GetParams(), convey.ShouldResemble, []interface{}{
			"pepp", "small", "medium", "hawai", "xlarge",
		})

		db = &DB{}
		db.stmt = newStmt(db)
		tx = db.Where(
			db.Where("pizza = ?", "pepp").Where(
				db.Where("size = ?", "small").Not("size = ?", "medium"),
			),
		).Or(
			db.Where("pizza = ?", "hawai").Where("size = ?", "xlarge"),
		)
		tx.stmt.setWhereClause()

		if tx.err != nil {
			t.Fatal(tx.err)
		}

		convey.So(tx.stmt.css[clause.KindWhere].GetContentWithPlaceHolder(), convey.ShouldEqual,
			`((pizza = ?) AND ((size = ?) AND (NOT (size = ?)))) OR ((pizza = ?) AND (size = ?))`)
		convey.So(tx.stmt.css[clause.KindWhere].GetParams(), convey.ShouldResemble, []interface{}{
			"pepp", "small", "medium", "hawai", "xlarge",
		})



	})
}

func TestDB_Not(t *testing.T) {
	convey.Convey("", t, func(){

		// NOT
		db := &DB{}
		db.stmt = newStmt(db)
		tx := db.Not("size = ?", "medium")
		tx.stmt.setWhereClause()

		if tx.err != nil {
			t.Fatal(tx.err)
		}

		convey.So(tx.stmt.css[clause.KindWhere].GetContentWithPlaceHolder(), convey.ShouldEqual,
			`NOT (size = ?)`)
		convey.So(tx.stmt.css[clause.KindWhere].GetParams(), convey.ShouldResemble, []interface{}{
			"medium",
		})

		db = &DB{}
		db.stmt = newStmt(db)

		tx = db.Not(&person{Name: "xxx"})
		tx.stmt.setWhereClause()

		convey.So(tx.stmt.css[clause.KindWhere].GetContentWithPlaceHolder(), convey.ShouldEqual,
			`NOT (Name = ?)`)
		convey.So(tx.stmt.css[clause.KindWhere].GetParams(), convey.ShouldResemble, []interface{}{
			"xxx",
		})

		db = &DB{}
		db.stmt = newStmt(db)
		tx = db.Not(&person{Name: "xxx"}, []string{"Name"}, []string{"Age"})
		tx.stmt.setWhereClause()

		convey.So(tx.stmt.css[clause.KindWhere].GetContentWithPlaceHolder(), convey.ShouldEqual,
			`NOT (Name = ? AND Age = ?)`)
		convey.So(tx.stmt.css[clause.KindWhere].GetParams(), convey.ShouldResemble, []interface{}{
			"xxx", (*uint16)(nil),
		})

		db = &DB{}
		db.stmt = newStmt(db)

		tx = db.Not(map[string]interface{}{"name": "xxx", "age": 10})
		tx.stmt.setWhereClause()

		convey.So(tx.stmt.css[clause.KindWhere].GetContentWithPlaceHolder(), convey.ShouldEqual,
			`NOT (name = ? AND age = ?)`)
		convey.So(tx.stmt.css[clause.KindWhere].GetParams(), convey.ShouldResemble, []interface{}{
			"xxx", 10,
		})


		// Group

		// Not And Or

		db = &DB{}
		db.stmt = newStmt(db)
		tx = db.Not(map[string]interface{}{"name": "xxx", "age": 10}).Or(map[string]interface{}{"gender": "male"})
		tx.stmt.setWhereClause()

		convey.So(tx.stmt.css[clause.KindWhere].GetContentWithPlaceHolder(), convey.ShouldEqual,
			`(NOT (name = ? AND age = ?)) OR (gender = ?)`)
		convey.So(tx.stmt.css[clause.KindWhere].GetParams(), convey.ShouldResemble, []interface{}{
			"xxx", 10, "male",
		})
	})
}


