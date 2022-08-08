package tls

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"net"
	"net/mail"
	"net/url"
	"os"
	"os/user"
	"strings"
	"time"
)

type CACert struct {
	Cert *x509.Certificate
	Key  crypto.PrivateKey
}

func userAndHostname() string {
	var uh string
	u, err := user.Current()
	if err == nil {
		uh = u.Username + "@"
	}
	if h, err := os.Hostname(); err == nil {
		uh += h
	}
	if err == nil && u.Name != "" && u.Name != u.Username {
		uh += " (" + u.Name + ")"
	}
	return uh
}

func GenerateKey(root bool) (*rsa.PrivateKey, error) {
	if root {
		return rsa.GenerateKey(rand.Reader, 4096)
	}
	return rsa.GenerateKey(rand.Reader, 2048)
}

// GenerateRoot return cert, privDER, nil
func GenerateRoot() (*x509.Certificate, *rsa.PrivateKey, error) {
	priv, _ := GenerateKey(true)
	pub := priv.Public()

	spkiASN1, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return nil, nil, err
	}

	var spki struct {
		Algorithm        pkix.AlgorithmIdentifier
		SubjectPublicKey asn1.BitString
	}
	_, err = asn1.Unmarshal(spkiASN1, &spki)
	if err != nil {
		return nil, nil, err
	}

	skid := sha1.Sum(spki.SubjectPublicKey.Bytes)

	tpl := &x509.Certificate{
		SerialNumber: serialNumber(),
		Subject: pkix.Name{
			Organization: []string{"BJCA"},
			CommonName:   "Root CA",
		},
		SubjectKeyId: skid[:],

		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(100, 0, 0),

		KeyUsage:    x509.KeyUsageCertSign,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},

		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLenZero:        true,
	}

	cert, err := x509.CreateCertificate(rand.Reader, tpl, tpl, pub, priv) //root cert
	if err != nil {
		return nil, nil, err
	}
	rootCert, err := x509.ParseCertificate(cert)
	if err != nil {
		return nil, nil, err
	}

	//privDER, err := x509.MarshalPKCS8PrivateKey(priv) //key
	//if err != nil {
	//	return nil, nil, err
	//}

	return rootCert, priv, nil
}

func ReadRootCertFile(filename string) (*x509.Certificate, error) {
	certPEMBlock, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	certDERBlock, _ := pem.Decode(certPEMBlock)
	if certDERBlock == nil || certDERBlock.Type != "CERTIFICATE" {
		return nil, errors.New("ERROR: failed to read the CA certificate: unexpected content")
	}
	rootCert, err := x509.ParseCertificate(certDERBlock.Bytes)
	if err != nil {
		return nil, err
	}
	return rootCert, err
}

//func ReadRootKeyFile(filename string) (interface{}, error) {
//	keyPEMBlock, err := ioutil.ReadFile(filename)
//	if err != nil {
//		return nil, err
//	}
//	keyDERBlock, _ := pem.Decode(keyPEMBlock)
//	if keyDERBlock == nil || keyDERBlock.Type != "PRIVATE KEY" {
//		return nil, errors.New("ERROR: failed to read the CA key: unexpected content")
//	}
//	rootKey, err := x509.ParsePKCS8PrivateKey(keyDERBlock.Bytes)
//	if err != nil {
//		return nil, err
//	}
//	return rootKey, err
//}

func CertificateInfo(certPath string) (*x509.Certificate, error) {
	certFile, err := ioutil.ReadFile(certPath)
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

func ReadPrivKeyFile(filename string) (*rsa.PrivateKey, error) {
	keyPEMBlock, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	keyDERBlock, _ := pem.Decode(keyPEMBlock)
	rootKey, err := x509.ParsePKCS1PrivateKey(keyDERBlock.Bytes)
	if err != nil {
		return nil, err
	}
	return rootKey, err
}

func ReadRootCert(cert []byte) (*x509.Certificate, error) {
	certDERBlock, _ := pem.Decode(cert)
	if certDERBlock == nil || certDERBlock.Type != "CERTIFICATE" {
		return nil, errors.New("ERROR: failed to read the CA certificate: unexpected content")
	}
	rootCert, err := x509.ParseCertificate(certDERBlock.Bytes)
	if err != nil {
		return nil, err
	}
	return rootCert, err
}

func ReadPrivKey(key []byte) (*rsa.PrivateKey, error) {
	keyDERBlock, _ := pem.Decode(key)
	rootKey, err := x509.ParsePKCS1PrivateKey(keyDERBlock.Bytes)
	if err != nil {
		return nil, err
	}
	return rootKey, err
}

// GenerateServer return certPEM, privPEM, nil
func (c CACert) GenerateServer(hosts []string) ([]byte, []byte, error) {
	priv, _ := GenerateKey(false)
	pub := priv.Public()

	keyID, err := hashPublicKey(&priv.PublicKey)
	if err != nil {
		return nil, nil, err
	}

	tpl := &x509.Certificate{
		SerialNumber: serialNumber(),
		Subject: pkix.Name{
			Organization:       []string{"CertAide"},
			OrganizationalUnit: []string{userAndHostname()},
			Country:            []string{"CN"},
			Locality:           []string{"BeiJing"},
			Province:           []string{"BeiJing"}, // S=
		},
		BasicConstraintsValid: true,
		IsCA:                  false,
		MaxPathLenZero:        false,
		SubjectKeyId:          keyID[:],
		NotBefore:             time.Now().AddDate(0, -1, 0), // 取当前时间存在与测试机的时效性
		NotAfter:              time.Now().AddDate(1, -1, 0),
		//x509.KeyUsageContentCommitment 防抵赖项不支持早期版本IE11
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			tpl.IPAddresses = append(tpl.IPAddresses, ip)
			if IsIPv4(h) && !IsSameIP(h, "127.0.0.1") && tpl.Subject.CommonName == "" {
				// IIS (the main target of PKCS #12 files), only shows the deprecated
				// Common Name in the UI. See issue #115.
				tpl.Subject.CommonName = h
			}
		} else if email, err := mail.ParseAddress(h); err == nil && email.Address == h {
			tpl.EmailAddresses = append(tpl.EmailAddresses, h)
		} else if uriName, err := url.Parse(h); err == nil && uriName.Scheme != "" && uriName.Host != "" {
			tpl.URIs = append(tpl.URIs, uriName)
		} else {
			tpl.DNSNames = append(tpl.DNSNames, h)
		}
	}

	if tpl.Subject.CommonName == "" {
		tpl.Subject.CommonName = hosts[0]
	}
	cert, err := x509.CreateCertificate(rand.Reader, tpl, c.Cert, pub, c.Key)
	if err != nil {
		return nil, nil, err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert})
	privDER, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return nil, nil, err
	}
	privPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privDER})

	return certPEM, privPEM, nil

}

func IsIP(host string) bool {
	return net.ParseIP(host) != nil
}

func IsSameIP(ipFirst, ipSecond string) bool {
	ip1 := net.ParseIP(ipFirst)
	ip2 := net.ParseIP(ipSecond)
	if ip1 == nil || ip2 == nil {
		return false
	}

	return ip1.Equal(ip2)
}

func IsIPv4(ip string) bool {
	if len(ip) == 0 {
		return false
	}

	if len(strings.Split(ip, ".")) != 4 {
		return false
	}

	return IsIP(ip)
}

func IsIPv6(ip string) bool {
	fmtIP, err := FormatIp(ip)
	if err != nil {
		return false
	}

	if len(fmtIP) == 0 {
		return false
	}

	if len(strings.Split(fmtIP, ":")) < 3 {
		return false
	}

	return IsIP(fmtIP)
}

func FormatIp(ipStr string) (string, error) {
	ip := net.ParseIP(strings.TrimSpace(ipStr))
	if ip == nil {
		return "", fmt.Errorf("IP地址错误")
	}
	return ip.String(), nil
}

func serialNumber() *big.Int {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	sn, _ := rand.Int(rand.Reader, serialNumberLimit)
	return sn
}

func WritePEM(filepath string, pem []byte) error {
	return ioutil.WriteFile(filepath, pem, 0644)
}

func hashPublicKey(key *rsa.PublicKey) ([]byte, error) {
	b, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return nil, fmt.Errorf("unable to hash key: %s", err)
	}

	h := sha1.New()
	h.Write(b)
	return h.Sum(nil), nil
}
