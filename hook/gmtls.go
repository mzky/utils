package hook

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/soheilhy/cmux"
	"github.com/tjfoc/gmsm/gmtls"
	"io"
	"net"
	"net/http"
)

// Config 服务器配置结构体
type Config struct {
	RSACertFile     string
	RSAKeyFile      string
	SM2SignCertFile string
	SM2SignKeyFile  string
	SM2EncCertFile  string
	SM2EncKeyFile   string
}

// buildConfig 加载 RSA/SM2 证书并构造支持 GM/RSA 自动切换的 gmtls.Config
// 通过 GetConfigForClient 为国密客户端和标准 TLS 客户端分别设置独立的算法参数。
func (c *Config) buildConfig() (*gmtls.Config, error) {
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

	// 基础混合配置：根据 ClientHello 版本自动选择 SM2 签名/加密证书或 RSA 证书
	baseConfig, err := gmtls.NewBasicAutoSwitchConfig(&sm2SignCert, &sm2EncCert, &rsaCert)
	if err != nil {
		return nil, fmt.Errorf("new auto switch config: %w", err)
	}

	// 在握手前返回针对当前协议的独立配置,实现 GM/RSA 独立算法参数
	baseConfig.GetConfigForClient = func(hello *gmtls.ClientHelloInfo) (*gmtls.Config, error) {
		if isGMSSL(hello) {
			cfg := baseConfig.Clone()
			cfg.GetKECertificate = baseConfig.GetKECertificate
			cfg.GMSupport = gmtls.NewGMSupport()
			cfg.Certificates = []gmtls.Certificate{sm2SignCert, sm2EncCert}
			cfg.CipherSuites = []uint16{
				gmtls.GMTLS_SM2_WITH_SM4_SM3,
				gmtls.GMTLS_ECDHE_SM2_WITH_SM4_SM3,
			}
			cfg.MinVersion = gmtls.VersionGMSSL
			cfg.MaxVersion = gmtls.VersionGMSSL
			cfg.InsecureSkipVerify = true
			return cfg, nil
		}

		cfg := baseConfig.Clone()
		cfg.GMSupport = nil
		cfg.Certificates = []gmtls.Certificate{rsaCert}
		cfg.CipherSuites = []uint16{
			gmtls.TLS_RSA_WITH_AES_128_GCM_SHA256,
			gmtls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		}
		cfg.MinVersion = gmtls.VersionTLS12
		cfg.MaxVersion = gmtls.VersionTLS12
		cfg.InsecureSkipVerify = true
		return cfg, nil
	}

	return baseConfig, nil
}

func isGMSSL(hello *gmtls.ClientHelloInfo) bool {
	for _, v := range hello.SupportedVersions {
		if v == gmtls.VersionGMSSL {
			return true
		}
	}
	return false
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

	httpsServer := &http.Server{Handler: handler}
	httpServer := &http.Server{Handler: redirectHandler()}

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
