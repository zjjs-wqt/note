package reuint

import (
	"testing"
)

func TestCopyDir(t *testing.T) {

	tests := []struct {
		name     string
		srcPath  string
		destPath string
		wantErr  bool
	}{
		{"复制", "C:\\development\\dpm\\target\\pubArea\\test\\1", "C:\\development\\dpm\\target\\pubArea\\test\\2", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//CopyDir(tt.srcPath, tt.destPath)
			if err := CopyDir(tt.srcPath, tt.destPath); err != nil {
				t.Errorf("CopyDir() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
