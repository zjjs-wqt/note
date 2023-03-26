package reuint

import (
	"fmt"
	"testing"
)

func TestPasswordProcess_Gen(t *testing.T) {

	pwdSaltHex, saltHex, err := GenPasswordSalt("123456")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(pwdSaltHex, saltHex)

	//step := map[string][]string{"admin": {"DELETE/api/logout"}}
	//fmt.Println(step)
}

func TestPasswordProcess_Verify(t *testing.T) {
	pwd := "123456"
	pwdSaltHex := "ebb4cb79911b6c71937e3b0b5aa9de4732178f162429472a111f1850ed047b68"
	saltHex := "d4476be7a88b5f9054a5937575374c20"

	ok := VerifyPasswordSalt(pwd, pwdSaltHex, saltHex)
	if !ok {
		t.Fatalf("Expect Password right, but not")
	}
	pwd = "111111"
	ok = VerifyPasswordSalt(pwd, pwdSaltHex, saltHex)
	if ok {
		t.Fatalf("Expect Password not right, but it pass")
	}
}
