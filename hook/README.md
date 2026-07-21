


// RSA算法例子,调用server.go

```go
package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/mzky/utils/hook"
	"net/http"
)

func main() {
	engine := gin.New()
	engine.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello World!")
	})
	engine.GET("/login.html", func(c *gin.Context) {
		c.String(http.StatusOK, "login.html ok!")
	})

	// The default value of `hook` is "HTTP/1.1 302 Found\r\n\r\n<script>location.protocol='https:'</script>\r\n"
	// Restore the default values of the standard library ↓
	// srv.SetDefaultBadRequest("HTTP/1.0 400 Bad Request\r\n\r\nClient sent an HTTP request to an HTTPS server.\n")

	srv := hook.NewServer(":1234", engine)

	// custom response
	//srv.SetRedirectPath("/login.html")
	//body := "---- message ----"
	//srv.SetResponse(body, func(r *http.Response) {
	//	r.StatusCode = 400
	//	r.Status = http.StatusText(400)
	//	r.Header.Set("Content-Type", "text/html")
	//})

	// custom response 2
	//srv.SetRedirectPath("/login.html")
	//srv.SetResponse("", func(r *http.Response) {
	//	r.StatusCode = 307
	//	r.Status = http.StatusText(307)
	//	r.Header.Set("Timing-Allow-Origin", "*")
	//	r.Header.Set("Non-Authoritative-Reason", "HSTS") // 这种方式重定向比script快
	//	r.Header.Set("x-xss-protection", "0")
	//	r.Header.Set("Cross-Origin-Resource-Policy", "Cross-Origin")
	//	r.Header.Set("Content-Type", "text/html; charset=utf-8")
	//	r.Header.Set("Location", "https://192.168.0.188:7569") // 可选配重定向地址
	//})

	fmt.Println(srv.ListenAndServeTLS("server.pem", "server.key"))
}
```


// 国密算法示例，调用gotlcp.go
```
// 指定tls版本和算法，剔除不安全的算法(漏扫原因)
		hook.TlsConfig = &tls.Config{
			NextProtos:         []string{"http/1.1"},
			InsecureSkipVerify: true,
			MaxVersion:         tls.VersionTLS13,
			MinVersion:         tls.VersionTLS12,
			CipherSuites: []uint16{
				// Tls 1.3
				tls.TLS_AES_128_GCM_SHA256,
				tls.TLS_AES_256_GCM_SHA384,
				tls.TLS_CHACHA20_POLY1305_SHA256,
				// Tls 1.2
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
				tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
			},
		}
		hook.TlcpConfig = &tlcp.Config{
            InsecureSkipVerify: true,
        }
		ln, err := hook.BuildListener(
            k.Config.TlsMode, // tlsMode包含gm，rsa可以指定算法和auto自动算法
			k.Config.ListenAddr,
			k.Config.ServerPem,
			k.Config.ServerKey,
			k.Config.GmSignCert,
			k.Config.GmSignKey,
			k.Config.GmEncCert,
			k.Config.GmEncKey,
		)
		if err != nil {
			logrus.Fatalln(err)
		}
		defer ln.Close()

		server := &http.Server{
			Addr:    k.Config.ListenAddr,
			Handler: k.Engine,
		}

		if e := server.Serve(ln); e != nil && !errors.Is(e, http.ErrServerClosed) {
			logrus.Fatalf("启动失败,查看端口是否被占用: %v", e)
		}
```

// 国密算法示例，调用gmtls.go
```
package main

import (
	"crypto/tls"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/tjfoc/gmsm/gmtls"

	"autotls/handler"
	"autotls/hook"
)

func main() {
	log.Println("Starting GM/TLS HTTP/HTTPS Service...")
	r := gin.Default()
	pd := &hook.ProtocolDetector{
		Handler: r,
		TLSConfig: &gmtls.Config{
			InsecureSkipVerify:       true,
			PreferServerCipherSuites: true,
			CipherSuites: []uint16{
				tls.TLS_AES_128_GCM_SHA256,
				tls.TLS_AES_256_GCM_SHA384,
				tls.TLS_CHACHA20_POLY1305_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
				tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
				// 保留最基础的国密加密套件
				gmtls.GMTLS_ECDHE_SM2_WITH_SM4_SM3,
				gmtls.GMTLS_SM2_WITH_SM4_SM3,
			},
			MinVersion: gmtls.VersionGMSSL,
			MaxVersion: tls.VersionTLS13,
		},
	}

	tlsConfig, err := pd.BuildTLSConfig(
		"auto",
		"./certs/server_sign.crt",
		"./certs/server_sign.key",
		"./certs/server_enc.crt",
		"./certs/server_enc.key",
		"./certs/server_std.crt",
		"./certs/server_std.key",
	)
	if err != nil {
		log.Fatalf("Failed to build TLS config: %v", err)
	}

	r.GET("/", handler.Index)
	r.GET("/health", handler.HealthCheck)

	if err := hook.ListenAndServe(":7569", tlsConfig, r); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
```