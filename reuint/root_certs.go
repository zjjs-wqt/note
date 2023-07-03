package reuint

import (
	"github.com/emmansun/gmsm/smx509"
	"note/appconf/dir"
	"os"
	"path/filepath"
)

var (
	CertPool *smx509.CertPool // 根证书池
)

// LoadCertsPool 初始化证书池 将根证书加入到证书池
func LoadCertsPool() {
	CertPool = smx509.NewCertPool()
	base, _ := filepath.Abs(dir.RootCertDir)
	// 读取文件夹
	readDir, err := os.ReadDir(base)
	if err != nil {
		return
	}
	// 根证书加入证书池
	for _, file := range readDir {
		path := filepath.Join(base, file.Name())
		temp, _ := os.ReadFile(path)
		cert, _ := smx509.ParseCertificate(Decode2DER(temp))
		CertPool.AddCert(cert)
	}
}
