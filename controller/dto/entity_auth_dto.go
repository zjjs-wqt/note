package dto

type EntityAuthDto struct {
	Ra        string `json:"Ra"`        // 随机数Ra
	Rb        string `json:"Rb"`        // 随机数Rb
	B         string `json:"B"`         // 可区分标识符B
	Text3     string `json:"text3"`     // 用户名
	Signature string `json:"signature"` // 签名值
}

type CertBindingDto struct {
	Cert      string `json:"cert"`      // 证书
	Ra        string `json:"Ra"`        // 随机数Ra
	Rb        string `json:"Rb"`        // 随机数Rb
	B         string `json:"B"`         // 可区分标识符B
	Text3     string `json:"text3"`     // 用户名
	Signature string `json:"signature"` // 签名值
}
