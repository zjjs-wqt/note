package appconf

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"testing"
)

func TestLoad(t *testing.T) {
	m := map[string][]string{
		"usr": []string{
			"GET/api/path",
			"POST/api/path2",
			"DELETE/api/path3",
		},

		"adm": []string{
			"/api/path",
			"/api/path2",
		},
	}
	actual, err := yaml.Marshal(m)
	if err != nil {
		//panic(err)
	}
	expect := `adm:
- /api/path
- /api/path2
usr:
- GET/api/path
- POST/api/path2
- DELETE/api/path3`
	if expect != string(actual) {
		t.Fatalf("ACL not match expect")
	}

	fmt.Printf("%s\n", expect)
}
