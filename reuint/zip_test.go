package reuint

import (
	"bytes"
	"testing"
)

func TestZip(t *testing.T) {

	buffer := bytes.NewBuffer([]byte{})
	err := Zip(buffer, "jwt", "filename.go")
	if err != nil {
		t.Fatal(err)
	}

}
