package tls

import (
	"crypto/rand"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"net"
	"os"
	"time"

	"github.com/tjfoc/gmsm/sm2"
	"github.com/tjfoc/gmsm/x509"
)

// GmCACert 国密CA证书结构体
type GmCACert struct {
	Cert  *x509.Certificate
	Key   *sm2.PrivateKey
	Hosts []string // 证书支持的主机名列表
}

// GenerateGmRoot 生成国密根证书
func GenerateGmRoot() (*GmCACert, error) {
	priv, err := sm2.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	tmpl := x509.Certificate{
		SerialNumber: SerialNumber(),
		Subject: pkix.Name{
			Organization:       []string{"BJCA"},
			CommonName:         "GM Root CA",
			OrganizationalUnit: []string{UserAndHostname()},
			Country:            []string{"CN"},
			Locality:           []string{"BeiJing"},
			Province:           []string{"BeiJing"},
		},
		NotBefore:             time.Now().AddDate(0, -1, 0),
		NotAfter:              time.Now().AddDate(100, 0, 0),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLenZero:        false,
	}

	certDER, err := x509.CreateCertificate(&tmpl, &tmpl, &priv.PublicKey, priv)
	if err != nil {
		return nil, err
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, err
	}

	return &GmCACert{
		Cert: cert,
		Key:  priv,
	}, nil
}

// GenKeyAndCert 生成密钥和证书
// types 0 为签名证书，1 为加密证书
func GenKeyAndCert(keyFile, certFile string, types int) error {
	priv, err := sm2.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}
	keyOut, err := os.Create(keyFile)
	if err != nil {
		return err
	}
	defer keyOut.Close()
	privBytes, err := x509.WritePrivateKeyToPem(priv, nil)
	if err != nil {
		return err
	}
	keyOut.Write(privBytes)
	var keyUsage x509.KeyUsage
	var extKeyUsage []x509.ExtKeyUsage

	// 0 为签名证书，1 为加密证书
	switch types {
	case 0:
		keyUsage = x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment
		extKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}
	case 1:
		keyUsage = x509.KeyUsageKeyAgreement | x509.KeyUsageDataEncipherment
		extKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageEmailProtection}
	}

	tmpl := x509.Certificate{
		SerialNumber: SerialNumber(),
		Subject: pkix.Name{
			Organization:       []string{"BJCA"},
			CommonName:         "Root CA Encryption",
			OrganizationalUnit: []string{UserAndHostname()},
			Country:            []string{"CN"},
			Locality:           []string{"BeiJing"},
			Province:           []string{"BeiJing"},
		},
		NotBefore:             time.Now().AddDate(0, -1, 0), // 取当前时间存在与测试机的时效性
		NotAfter:              time.Now().AddDate(1, 0, 0),
		KeyUsage:              keyUsage,
		ExtKeyUsage:           extKeyUsage,
		BasicConstraintsValid: true,
	}
	certDER, err := x509.CreateCertificate(&tmpl, &tmpl, &priv.PublicKey, priv)
	if err != nil {
		return err
	}
	certOut, err := os.Create(certFile)
	if err != nil {
		return err
	}
	defer certOut.Close()
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	return nil
}

func ReadGmRootCert(cert []byte) (*x509.Certificate, error) {
	return x509.ReadCertificateFromPem(cert)
}

func ReadGmPrivKey(key []byte) (*sm2.PrivateKey, error) {
	return x509.ReadPrivateKeyFromPem(key, nil)
}

// GenerateGmCert 使用国密根证书生成子证书
// types 0 为签名证书，1 为加密证书
func (c *GmCACert) GenerateGmCert(keyFile, certFile string, types int) error {
	priv, err := sm2.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}

	// 处理Hosts数组，添加DNS名称和IP地址
	if len(c.Hosts) < 1 {
		c.Hosts = append(c.Hosts, "127.0.0.1")
	}

	// 保存私钥
	keyOut, err := os.Create(keyFile)
	if err != nil {
		return err
	}
	defer keyOut.Close()
	privBytes, err := x509.WritePrivateKeyToPem(priv, nil)
	if err != nil {
		return err
	}
	keyOut.Write(privBytes)

	var keyUsage x509.KeyUsage
	var extKeyUsage []x509.ExtKeyUsage

	// 0 为签名证书，1 为加密证书
	switch types {
	case 0:
		keyUsage = x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment
		extKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}
	case 1:
		keyUsage = x509.KeyUsageKeyAgreement | x509.KeyUsageDataEncipherment
		extKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageEmailProtection}
	}

	tmpl := x509.Certificate{
		SerialNumber: SerialNumber(),
		Subject: pkix.Name{
			Organization:       []string{"BJCA"},
			CommonName:         c.Hosts[0],
			OrganizationalUnit: []string{UserAndHostname()},
			Country:            []string{"CN"},
			Locality:           []string{"BeiJing"},
			Province:           []string{"BeiJing"},
		},
		NotBefore:             time.Now().AddDate(0, -1, 0),
		NotAfter:              time.Now().AddDate(1, 0, 0),
		KeyUsage:              keyUsage,
		ExtKeyUsage:           extKeyUsage,
		BasicConstraintsValid: true,
		IsCA:                  false,
	}

	for _, host := range c.Hosts {
		if ip := net.ParseIP(host); ip != nil {
			tmpl.IPAddresses = append(tmpl.IPAddresses, ip)
		} else {
			tmpl.DNSNames = append(tmpl.DNSNames, host)
		}
	}

	// 使用根证书签名生成子证书
	certDER, err := x509.CreateCertificate(&tmpl, c.Cert, &priv.PublicKey, c.Key)
	if err != nil {
		return err
	}

	// 保存证书
	certOut, err := os.Create(certFile)
	if err != nil {
		return err
	}
	defer certOut.Close()
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	return nil
}

func GmCertificateInfo(certPath string) (*x509.Certificate, error) {
	certFile, err := os.ReadFile(certPath)
	if err != nil {
		return nil, errors.New("地址或权限异常") // 创建第一个证书&异常情况创建证书
	}

	pemBlock, _ := pem.Decode(certFile)
	if pemBlock == nil {
		return nil, errors.New("证书格式错误")
	}

	cert, err := x509.ParseCertificate(pemBlock.Bytes)
	if err != nil {
		return nil, errors.New("证书解析异常")
	}
	return cert, nil
}

// SaveGmCertAndKey 保存国密证书和密钥到文件
func (c *GmCACert) SaveGmCertAndKey(certFile, keyFile string) error {
	// 保存证书
	certOut, err := os.Create(certFile)
	if err != nil {
		return err
	}
	defer certOut.Close()
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: c.Cert.Raw})

	// 保存密钥
	keyOut, err := os.Create(keyFile)
	if err != nil {
		return err
	}
	defer keyOut.Close()
	keyBytes, err := x509.WritePrivateKeyToPem(c.Key, nil)
	if err != nil {
		return err
	}
	keyOut.Write(keyBytes)
	return nil
}
