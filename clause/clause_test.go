package clause

import (
	"github.com/smartystreets/goconvey/convey"
	"testing"
)

func Test_cond_setQueryAndParams(t *testing.T) {
	convey.Convey("", t, func() {
		cond1 := &cond{
			kind:                 CondKindWhere,
			queryWithPlaceHolder: `field1 = ?`,
			params:               []interface{}{"val1"},
		}

		cond2 := &cond{
			kind:                 CondKindOr,
			queryWithPlaceHolder: `field2 = ?`,
			params:               []interface{}{2},
		}

		cond3 := &cond{
			kind:                 CondKindOr,
			queryWithPlaceHolder: `field3 = ?`,
			params:               []interface{}{true},
		}

		cond4 := &cond{
			kind:    CondKindWhere,
			children: []*cond{cond1, cond2},
		}

		cond5 := &cond{
			kind:     CondKindWhere,
			children: []*cond{cond4, cond3},
		}

		cond5.setQueryAndParams()

		convey.So(cond5.queryWithPlaceHolder, convey.ShouldEqual,
			`((field1 = ?) OR (field2 = ?)) OR (field3 = ?)`)
		convey.So(cond5.params, convey.ShouldResemble, []interface{}{
			"val1", 2, true,
		})
	})
}

func Test_buildWhereCondByMap(t *testing.T) {
	convey.Convey("", t, func() {
		m := map[string]interface{}{
			"field1": "val1",
			"field2": "val2",
		}

		cd := BuildCondByMap(m, CondKindWhere)
		cd.setQueryAndParams()

		convey.So(cd.queryWithPlaceHolder, convey.ShouldEqual,
			`field1 = ? AND field2 = ?`)
		convey.So(cd.params, convey.ShouldResemble,
			[]interface{}{
				"val1", "val2",
			})
	})
}

func Test_buildWhereCondByStruct(t *testing.T) {
	convey.Convey("", t, func() {
		obj := &struct {
			Name string
			Age  int
			Male bool
		}{
			Name: "xxx",
			Age:  18,
		}

		cd, err := BuildCondByStruct(obj, CondKindWhere)
		convey.So(err, convey.ShouldBeNil)
		cd.setQueryAndParams()

		convey.So(cd.queryWithPlaceHolder, convey.ShouldEqual,
			`Name = ? AND Age = ?`)
		convey.So(cd.params, convey.ShouldResemble,
			[]interface{}{
				"xxx", 18,
			})

		cd, err = BuildCondByStruct(obj, CondKindWhere, []string{"Name"}, []string{"Male"})
		convey.So(err, convey.ShouldBeNil)
		cd.setQueryAndParams()

		convey.So(cd.queryWithPlaceHolder, convey.ShouldEqual,
			`Name = ? AND Male = ?`)
		convey.So(cd.params, convey.ShouldResemble,
			[]interface{}{
				"xxx", false,
			})

		cd, err = BuildCondByStruct(obj, CondKindWhere, "Male", "Age")
		convey.So(err, convey.ShouldBeNil)
		cd.setQueryAndParams()

		convey.So(cd.queryWithPlaceHolder, convey.ShouldEqual,
			`Male = ? AND Age = ?`)
		convey.So(cd.params, convey.ShouldResemble,
			[]interface{}{
				false, 18,
			})

	})
}
