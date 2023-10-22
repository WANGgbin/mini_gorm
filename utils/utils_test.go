package utils

import (
	"github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestTransFromHumpToSnake(t *testing.T) {
	convey.Convey("", t, func() {
		testCases := []struct {
			name string
			want string
		}{
			{
				name: "ANameBeginWithA",
				want: "a_name_begin_with_a",
			},
			{
				name: "person",
				want: "person",
			},
			{
				name: "aNameBeginWithA",
				want: "a_name_begin_with_a",
			},
			{
				name: "PERSON",
				want: "p_e_r_s_o_n",
			},
		}

		for _, testCase := range testCases {
			convey.So(testCase.want, convey.ShouldEqual, TransFromHumpToSnake(testCase.name))
		}
	})
}