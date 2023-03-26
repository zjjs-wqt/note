package reuint

import (
	"reflect"
	"testing"
)

func TestStrToIntSlice(t *testing.T) {

	tests := []struct {
		name string
		args string
		want []int
	}{
		{"CASE 1", ",,", []int{0, 0, 0}},
		{"CASE 2", "1,A,C", []int{1, 0, 0}},
		{"CASE 3", "[1,A,C]", []int{0, 0, 0}},
		{"CASE 4", "", []int{0}},
		{"CASE 5", "26,27", []int{26, 27}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StrToIntSlice(tt.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StrToIntSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}
