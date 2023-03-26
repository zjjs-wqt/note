package reuint

import (
	"log"
	"net/url"
	"testing"
)

func TestVerifyToken(t *testing.T) {
	token := "eyJhbGciOiJITUFDLVNNMyIsInR5cCI6IkpXVCJ9.eyJ0eXBlIjoidXNlciIsInN1YiI6MSwiZXhwIjoxNjc2NTU2OTU1OTExLCJjbGllbnRfaWQiOiIifQ%3D%3D.CiC9EM2-apypwTjvFwvB9spl4ngTF8G2FulH-6gmZ_o%3D"
	accessToken, _ := url.QueryUnescape(token)
	claims, err := VerifyToken(accessToken)
	if err != nil {
		t.Fatal(err)
	}
	log.Println(claims)
}
