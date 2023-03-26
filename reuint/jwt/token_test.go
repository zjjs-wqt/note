package jwt

import (
	"fmt"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	key := make([]byte, 32)
	claims := &Claims{Exp: time.Now().UnixMilli() - 1000}
	token := New(key, claims)
	cc, err := Verify(key, token)
	if err == nil {
		t.Fatal("token should expired, but not")
	}

	claims = &Claims{Exp: time.Now().UnixMilli() + 1000}
	token = New(key, claims)
	cc, err = Verify(key, token)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(token)
	fmt.Printf("%+v\n", cc)
}
