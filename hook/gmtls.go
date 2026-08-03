package hook

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/soheilhy/cmux"
	"github.com/tjfoc/gmsm/gmtls"
)

const (
	ModeAuto = "auto"
	ModeGM   = "gm"
	ModeRSA  = "rsa"
)

// 默认 GM 密码套件
var defaultGMCipherSuites = []uint16{
	gmtls.GMTLS_SM2_WITH_SM4_SM3,
	gmtls.GMTLS_ECDHE_SM2_WITH_SM4_SM3,
}

// 默认 RSA/TLS 密码套件
var defaultRSACipherSuites = []uint16{
	gmtls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
	gmtls.TLS_RSA_WITH_AES_128_GCM_SHA256,
}

// Config 服务器配置结构体
type Config struct {
	RSACertFile     string
	RSAKeyFile      string
	SM2SignCertFile string
	SM2SignKeyFile  string
	SM2EncCertFile  string
	SM2EncKeyFile   string

	// Mode 算法模式: auto, gm, rsa
	Mode string
	// CipherSuites 自定义密码套件（可选），nil 时使用各模式默认值
	CipherSuites []uint16

	Server *http.Server
}

// GetCipherSuites 返回当前模式应使用的密码套件
func (c *Config) GetCipherSuites() []uint16 {
	if c.CipherSuites != nil {
		return c.CipherSuites
	}
	switch c.Mode {
	case ModeGM:
		return defaultGMCipherSuites
	case ModeRSA:
		return defaultRSACipherSuites
	default:
		// auto 模式下返回 GM + RSA 所有套件
		return append(append([]uint16{}, defaultGMCipherSuites...), defaultRSACipherSuites...)
	}
}

// buildConfig 根据 Mode 构造 gmtls.Config
func (c *Config) buildConfig() (*gmtls.Config, error) {
	switch c.Mode {
	case ModeGM:
		return c.buildGMConfig()
	case ModeRSA:
		return c.buildRSAConfig()
	default:
		return c.buildAutoConfig()
	}
}

// buildGMConfig 构建国密模式配置（仅 SM2 证书 + GM 算法套件）
func (c *Config) buildGMConfig() (*gmtls.Config, error) {
	sm2SignCert, err := gmtls.LoadX509KeyPair(c.SM2SignCertFile, c.SM2SignKeyFile)
	if err != nil {
		return nil, fmt.Errorf("load sm2 sign cert: %w", err)
	}

	sm2EncCert, err := gmtls.LoadX509KeyPair(c.SM2EncCertFile, c.SM2EncKeyFile)
	if err != nil {
		return nil, fmt.Errorf("load sm2 enc cert: %w", err)
	}

	return &gmtls.Config{
		GMSupport:    gmtls.NewGMSupport(),
		Certificates: []gmtls.Certificate{sm2SignCert, sm2EncCert},
		GetKECertificate: func(info *gmtls.ClientHelloInfo) (*gmtls.Certificate, error) {
			return &sm2EncCert, nil
		},
		CipherSuites: c.GetCipherSuites(),
		MinVersion:   gmtls.VersionGMSSL,
		MaxVersion:   gmtls.VersionGMSSL,
	}, nil
}

// buildRSAConfig 构建 RSA 模式配置（仅 RSA 证书 + TLS 算法套件）
func (c *Config) buildRSAConfig() (*gmtls.Config, error) {
	rsaCert, err := gmtls.LoadX509KeyPair(c.RSACertFile, c.RSAKeyFile)
	if err != nil {
		return nil, fmt.Errorf("load rsa cert: %w", err)
	}

	return &gmtls.Config{
		Certificates: []gmtls.Certificate{rsaCert},
		CipherSuites: c.GetCipherSuites(),
		MinVersion:   gmtls.VersionTLS12,
		MaxVersion:   gmtls.VersionTLS12,
	}, nil
}

// buildAutoConfig 构建自动切换模式配置（GM/RSA 自动协商）
func (c *Config) buildAutoConfig() (*gmtls.Config, error) {
	rsaCert, err := gmtls.LoadX509KeyPair(c.RSACertFile, c.RSAKeyFile)
	if err != nil {
		return nil, fmt.Errorf("load rsa cert: %w", err)
	}

	sm2SignCert, err := gmtls.LoadX509KeyPair(c.SM2SignCertFile, c.SM2SignKeyFile)
	if err != nil {
		return nil, fmt.Errorf("load sm2 sign cert: %w", err)
	}

	sm2EncCert, err := gmtls.LoadX509KeyPair(c.SM2EncCertFile, c.SM2EncKeyFile)
	if err != nil {
		return nil, fmt.Errorf("load sm2 enc cert: %w", err)
	}

	config, err := gmtls.NewBasicAutoSwitchConfig(&sm2SignCert, &sm2EncCert, &rsaCert)
	if err != nil {
		return nil, fmt.Errorf("new auto switch config: %w", err)
	}

	// 应用自定义密码套件（如果指定）
	config.CipherSuites = c.GetCipherSuites()

	// 允许 GMSSL 到 TLS1.2 的版本协商
	config.MinVersion = gmtls.VersionGMSSL
	config.MaxVersion = gmtls.VersionTLS12

	return config, nil
}

func newServerFromTemplate(template *http.Server, handler http.Handler) *http.Server {
	if template == nil {
		return &http.Server{
			Handler: handler,
		}
	}
	s := *template
	s.Handler = handler
	return &s
}

// StartMixedServer 在 addr 单一端口上同时监听 HTTPS（gmtls,支持 GM/RSA 自动切换）
// 与 HTTP,对所有 HTTP 请求返回 307 重定向到同地址 HTTPS。
func (c *Config) StartMixedServer(addr string, handler http.Handler) error {
	config, err := c.buildConfig()
	if err != nil {
		return fmt.Errorf("build config: %w", err)
	}

	ltr, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", addr, err)
	}

	m := cmux.New(ltr)

	// TLS/GMSSL 匹配：首字节 0x16 为握手记录,后两字节为协议版本
	tlsListener := m.Match(cmux.Matcher(isTLSRecord))
	// 其余流量均视为 HTTP,返回 307 重定向
	httpListener := m.Match(cmux.Any())

	gmtlsListener := gmtls.NewListener(tlsListener, config)

	httpsServer := newServerFromTemplate(c.Server, handler)
	httpServer := newServerFromTemplate(c.Server, redirectHandler())

	go func() {
		if e := httpsServer.Serve(gmtlsListener); e != nil && !errors.Is(e, cmux.ErrListenerClosed) {
			logrus.Debugf("https server error: %v", e)
		}
	}()

	go func() {
		if e := httpServer.Serve(httpListener); e != nil && !errors.Is(e, cmux.ErrListenerClosed) {
			logrus.Debugf("http server error: %v", e)
		}
	}()

	logrus.Debugf("Listening on %s (HTTP->HTTPS redirect, GM/RSA auto-switch)", addr)
	return m.Serve()
}

// isTLSRecord 判断 reader 前 3 字节是否为 TLS/GMSSL 记录层头部
func isTLSRecord(r io.Reader) bool {
	buf := make([]byte, 3)
	if _, err := io.ReadFull(r, buf); err != nil {
		return false
	}
	if buf[0] != 0x16 {
		return false
	}
	version := uint16(buf[1])<<8 | uint16(buf[2])
	switch version {
	case gmtls.VersionGMSSL, gmtls.VersionSSL30, gmtls.VersionTLS10,
		gmtls.VersionTLS11, gmtls.VersionTLS12:
		return true
	}
	return false
}

// redirectHandler 返回 307 Temporary Redirect,跳转到同源 HTTPS 地址
func redirectHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		target := "https://" + r.Host + r.URL.RequestURI()
		http.Redirect(w, r, target, http.StatusTemporaryRedirect)
	})
}
