package tls

import (
	"crypto/rand"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"time"

	"github.com/tjfoc/gmsm/sm2"
	"github.com/tjfoc/gmsm/x509"
)

// GenKeyAndCert 生成密钥和证书
// serial 为证书序列号
// types 0 为签名证书，1 为加密证书
func GenKeyAndCert(keyFile, certFile string, serial int64, types int) error {
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
		SerialNumber: big.NewInt(serial),
		Subject: pkix.Name{
			Organization:       []string{"BJCA"},
			CommonName:         "Root CA Encryption",
			OrganizationalUnit: []string{"BJCADevice"},
			Country:            []string{"CN"},
			Locality:           []string{"BeiJing"},
			Province:           []string{"BeiJing"},
		},
		NotBefore:             time.Now(),
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
