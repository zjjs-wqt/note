package reuint

import "testing"

func TestGetFileType(t *testing.T) {
	type args struct {
		fileType string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"Case .txt", args{fileType: ".txt"}, "text/plain"},
		{"Case exe", args{fileType: "exe"}, "application/octet-stream"},
		{"Case ..doc", args{fileType: "..doc"}, "application/octet-stream"},
		{"Case  ", args{fileType: " "}, "application/octet-stream"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetMIME(tt.args.fileType); got != tt.want {
				t.Errorf("GetMIME() = %v, want %v", got, tt.want)
			}
		})
	}
}
