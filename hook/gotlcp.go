package hook

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"sync"

	"gitee.com/Trisia/gotlcp/pa"
	"gitee.com/Trisia/gotlcp/tlcp"
)

func loadConfigs(mode, certFile, keyFile, signCert, signKey, encCert, encKey string) (*tls.Config, *tlcp.Config, error) {
	var tlsConfig *tls.Config
	if mode == "tls" || mode == "auto" {
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return nil, nil, fmt.Errorf("加载标准 TLS 证书: %w", err)
		}
		tlsConfig = &tls.Config{
			MinVersion:   tls.VersionTLS12,
			NextProtos:   []string{"http/1.1"},
			Certificates: []tls.Certificate{cert},
		}
	}

	var tlcpConfig *tlcp.Config
	if mode == "tlcp" || mode == "gm" || mode == "auto" {
		sign, err := tlcp.LoadX509KeyPair(signCert, signKey)
		if err != nil {
			return nil, nil, fmt.Errorf("加载 TLCP 签名证书: %w", err)
		}
		enc, err := tlcp.LoadX509KeyPair(encCert, encKey)
		if err != nil {
			return nil, nil, fmt.Errorf("加载 TLCP 加密证书: %w", err)
		}
		tlcpConfig = &tlcp.Config{Certificates: []tlcp.Certificate{sign, enc}}
	}

	return tlsConfig, tlcpConfig, nil
}

func displayHTTPSURL(addr string) string {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return "https://" + addr
	}
	if host == "" || host == "0.0.0.0" || host == "::" {
		host = "localhost"
	}
	return "https://" + net.JoinHostPort(host, port)
}

type httpRedirectListener struct {
	net.Listener
}

func newHTTPRedirectListener(inner net.Listener) net.Listener {
	return &httpRedirectListener{Listener: inner}
}

func (l *httpRedirectListener) Accept() (net.Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}
	return &httpRedirectConn{
		Conn:   conn,
		reader: bufio.NewReader(conn),
	}, nil
}

type httpRedirectConn struct {
	net.Conn
	reader     *bufio.Reader
	once       sync.Once
	redirected bool
	readErr    error
}

func (c *httpRedirectConn) Read(p []byte) (int, error) {
	c.once.Do(func() {
		first, err := c.reader.Peek(1)
		if err != nil {
			c.readErr = err
			return
		}
		if first[0] >= 'A' && first[0] <= 'Z' {
			c.redirected = true
			c.readErr = c.redirectHTTP()
		}
	})

	if c.redirected || c.readErr != nil {
		if c.readErr != nil {
			return 0, c.readErr
		}
		return 0, io.EOF
	}
	return c.reader.Read(p)
}

func (c *httpRedirectConn) redirectHTTP() error {
	req, err := http.ReadRequest(c.reader)
	if err != nil {
		_, _ = io.WriteString(c.Conn, "HTTP/1.1 400 Bad Request\r\nConnection: close\r\nContent-Length: 0\r\n\r\n")
		return io.EOF
	}
	defer req.Body.Close()

	host := req.Host
	if host == "" {
		host = c.LocalAddr().String()
	}
	location := (&url.URL{
		Scheme:   "https",
		Host:     host,
		Path:     req.URL.Path,
		RawPath:  req.URL.RawPath,
		RawQuery: req.URL.RawQuery,
	}).String()

	_, _ = fmt.Fprintf(c.Conn, "HTTP/1.1 307 Temporary Redirect\r\n")
	_, _ = fmt.Fprintf(c.Conn, "Location: %s\r\n", location)
	_, _ = fmt.Fprintf(c.Conn, "Connection: close\r\n")
	_, _ = fmt.Fprintf(c.Conn, "Content-Length: 0\r\n")
	_, _ = fmt.Fprintf(c.Conn, "Timing-Allow-Origin: *\r\n")
	_, _ = fmt.Fprintf(c.Conn, "Non-Authoritative-Reason: HSTS\r\n")
	_, _ = fmt.Fprintf(c.Conn, "x-xss-protection: 0\r\n")
	_, _ = fmt.Fprintf(c.Conn, "Cross-Origin-Resource-Policy: Cross-Origin\r\n")
	_, _ = fmt.Fprintf(c.Conn, "\r\n")

	return io.EOF
}

func (c *httpRedirectConn) Close() error {
	return c.Conn.Close()
}

func BuildListener(mode, addr, certFile, keyFile, signCert, signKey, encCert, encKey string) (net.Listener, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	ln = newHTTPRedirectListener(ln)

	tlsConfig, tlcpConfig, err := loadConfigs(mode, certFile, keyFile, signCert, signKey, encCert, encKey)
	if err != nil {
		return nil, err
	}

	switch mode {
	case "tls", "rsa":
		ln = tls.NewListener(ln, tlsConfig)
	case "tlcp", "gm":
		ln = tlcp.NewListener(ln, tlcpConfig)
	case "auto":
		ln = pa.NewListener(ln, tlcpConfig, tlsConfig)
	default:
		return nil, fmt.Errorf("不支持的模式 %q，可选值为 tls、tlcp(gm)、auto", mode)
	}

	return ln, nil
}
