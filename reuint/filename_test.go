package reuint

import (
	"fmt"
	"strings"
	"testing"
)

func TestGenFileName(t *testing.T) {
	fmt.Println(GenTimeFileName("a"))
	str := GenTimeFileName("a.txt")
	if !strings.HasSuffix(str, ".txt") {
		t.Fatal("expect get '.txt' suffix")
	}
	str = GenTimeFileName("")
	if str == "" {
		t.Fatal("expect get time base random name, but not")
	}
}
