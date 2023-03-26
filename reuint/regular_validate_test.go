package reuint

import (
	"testing"
)

func TestPhoneValidate(t *testing.T) {
	tests := []struct {
		name string
		str  string
		want bool
	}{
		{"CASE1", "123456", false},
		{"CASE2", "1385555555", false},
		{"CASE3", "", false},
		{"CASE4", "13844444444", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PhoneValidate(tt.str); got != tt.want {
				t.Errorf("PhoneValidate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEmailValidate(t *testing.T) {
	tests := []struct {
		name string
		str  string
		want bool
	}{
		{"CASE1", "@qq.com", false},
		{"CASE2", "123@qq", false},
		{"CASE3", "123@qq.com", true},
		{"CASE4", "123@qq.cn", true},
		{"CASE5", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EmailValidate(tt.str); got != tt.want {
				t.Errorf("EmailValidate() = %v, want %v", got, tt.want)
			}
		})
	}
}
