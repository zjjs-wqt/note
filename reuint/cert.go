package reuint

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"github.com/emmansun/gmsm/smx509"
	"strings"
)

// Decode2DER 解析编码并转码为DER格式
// data: 待解析的内容
// return DER编码
func Decode2DER(data []byte) []byte {
	if len(data) == 0 {
		return nil
	}
	// 尝试1 DER ASN1 Sequence 格式
	// Class: 0, P/C=C(1)  tag: 0x10 => 30 (SEQUENCE)
	if data[0] == 0x30 {
		return data
	}

	// 尝试2 PEM格式
	str := string(data)
	if strings.HasPrefix(str, "-----BEGIN") {
		block, _ := pem.Decode(data)
		if block != nil {
			return block.Bytes
		}
	}

	// 尝试3 BASE64 ASN1 Sequence
	if raw, err := base64.StdEncoding.DecodeString(str); err == nil {
		if raw[0] != 0x30 {
			return nil
		}
		return raw
	}

	//  尝试4 HexString 16进制字符串
	if bin, err := hex.DecodeString(str); err == nil {
		return bin
	}

	// 未知类型
	return nil
}

// ParseCert 解析base64编码的证书信息
// data: 待解析的证书（base64编码）
// return 证书信息, 错误
func ParseCert(data string) (*smx509.Certificate, error) {
	if len(data) <= 0 {
		return nil, nil
	}
	cert, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}
	certificate, err := smx509.ParseCertificate(Decode2DER(cert))
	if err != nil {
		return nil, err
	}
	return certificate, nil
}
