package reuint

import (
	"testing"
)

func TestPinyinConversion(t *testing.T) {

	tests := []struct {
		name    string
		str     string
		want    string
		wantErr bool
	}{
		{"CASE 1", "中国人ADS123", "zgr", false},
		{"CASE 2", "中as国sd人", "zgr", false},
		{"CASE 3", "中...国 人", "zgr", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := PinyinConversion(tt.str)
			if (err != nil) != tt.wantErr {
				t.Errorf("PinyinConversion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("PinyinConversion() got = %v, want %v", got, tt.want)
			}
		})
	}
}
