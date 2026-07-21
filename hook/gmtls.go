package hook

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/tjfoc/gmsm/gmtls"
)

func BuildTLSConfig(mode, sm2SignCert, sm2SignKey, sm2EncCert, sm2EncKey, stdCert, stdKey string) (*gmtls.Config, error) {
	switch mode {
	case "gm":
		return buildGMOnlyConfig(sm2SignCert, sm2SignKey, sm2EncCert, sm2EncKey)
	case "std", "rsa":
		return buildStdOnlyConfig(stdCert, stdKey)
	default:
		return buildAutoSwitchConfig(sm2SignCert, sm2SignKey, sm2EncCert, sm2EncKey, stdCert, stdKey)
	}
}

func buildAutoSwitchConfig(sm2SignCert, sm2SignKey, sm2EncCert, sm2EncKey, stdCert, stdKey string) (*gmtls.Config, error) {
	log.Println("TLS Mode: Auto Switch (GM + Standard)")

	sigCert, err := gmtls.LoadX509KeyPair(sm2SignCert, sm2SignKey)
	if err != nil {
		log.Printf("Warning: Failed to load SM2 sign certificate: %v", err)
		return buildStdOnlyConfig(stdCert, stdKey)
	}

	encCert, err := gmtls.LoadX509KeyPair(sm2EncCert, sm2EncKey)
	if err != nil {
		log.Printf("Warning: Failed to load SM2 encrypt certificate: %v", err)
		return buildStdOnlyConfig(stdCert, stdKey)
	}

	var stdCertPtr *gmtls.Certificate
	if stdCert != "" && stdKey != "" {
		cert, err := gmtls.LoadX509KeyPair(stdCert, stdKey)
		if err != nil {
			log.Printf("Warning: Failed to load standard certificate: %v", err)
		} else {
			stdCertPtr = &cert
		}
	}

	config, err := gmtls.NewBasicAutoSwitchConfig(&sigCert, &encCert, stdCertPtr)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func buildGMOnlyConfig(sm2SignCert, sm2SignKey, sm2EncCert, sm2EncKey string) (*gmtls.Config, error) {
	log.Println("TLS Mode: GM Only")

	sigCert, err := gmtls.LoadX509KeyPair(sm2SignCert, sm2SignKey)
	if err != nil {
		return nil, err
	}

	encCert, err := gmtls.LoadX509KeyPair(sm2EncCert, sm2EncKey)
	if err != nil {
		return nil, err
	}

	config := &gmtls.Config{
		GMSupport:    &gmtls.GMSupport{},
		Certificates: []gmtls.Certificate{sigCert, encCert},
	}

	return config, nil
}

func buildStdOnlyConfig(stdCert, stdKey string) (*gmtls.Config, error) {
	log.Println("TLS Mode: Standard Only")

	cert, err := gmtls.LoadX509KeyPair(stdCert, stdKey)
	if err != nil {
		return nil, err
	}

	config := &gmtls.Config{
		Certificates: []gmtls.Certificate{cert},
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		},
		MinVersion: tls.VersionTLS12,
	}

	return config, nil
}

type ProtocolDetector struct {
	TLSConfig *gmtls.Config
	Handler   http.Handler
}

func ListenAndServe(addr string, tlsConfig *gmtls.Config, handler http.Handler) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to create listener: %w", err)
	}
	log.Printf("Server starting on %s", addr)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("Shutting down server...")
		listener.Close()
	}()

	detector := &ProtocolDetector{
		TLSConfig: tlsConfig,
		Handler:   handler,
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Accept error: %v", err)
			return nil
		}
		go detector.HandleConnection(conn)
	}
}

func (pd *ProtocolDetector) HandleConnection(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 1)
	n, err := conn.Read(buf)
	if err != nil || n != 1 {
		return
	}

	if buf[0] == 0x16 {
		pc := &prefixedConn{prefix: buf, conn: conn}
		tlsConn := gmtls.Server(pc, pd.TLSConfig)
		if err := tlsConn.Handshake(); err != nil {
			return
		}
		bufConn := bufio.NewReader(tlsConn)
		for {
			req, err := http.ReadRequest(bufConn)
			if err != nil {
				break
			}
			w := &responseWriter{
				conn:    tlsConn,
				headers: make(http.Header),
			}
			pd.Handler.ServeHTTP(w, req)
			w.WriteHeaderIfNotWritten()
			w.WriteResponse()
			if req.Close {
				break
			}
		}
	} else {
		pd.handleHTTPRedirect(buf, conn)
	}
}

type prefixedConn struct {
	prefix []byte
	conn   net.Conn
}

func (pc *prefixedConn) Read(p []byte) (n int, err error) {
	if len(pc.prefix) > 0 {
		n = copy(p, pc.prefix)
		pc.prefix = pc.prefix[n:]
		return n, nil
	}
	return pc.conn.Read(p)
}

func (pc *prefixedConn) Write(p []byte) (n int, err error) {
	return pc.conn.Write(p)
}

func (pc *prefixedConn) Close() error {
	return pc.conn.Close()
}

func (pc *prefixedConn) LocalAddr() net.Addr {
	return pc.conn.LocalAddr()
}

func (pc *prefixedConn) RemoteAddr() net.Addr {
	return pc.conn.RemoteAddr()
}

func (pc *prefixedConn) SetDeadline(t time.Time) error {
	return pc.conn.SetDeadline(t)
}

func (pc *prefixedConn) SetReadDeadline(t time.Time) error {
	return pc.conn.SetReadDeadline(t)
}

func (pc *prefixedConn) SetWriteDeadline(t time.Time) error {
	return pc.conn.SetWriteDeadline(t)
}

type responseWriter struct {
	conn          net.Conn
	headers       http.Header
	statusCode    int
	body          []byte
	headerWritten bool
}

func (w *responseWriter) Header() http.Header {
	return w.headers
}

func (w *responseWriter) Write(p []byte) (n int, err error) {
	w.body = append(w.body, p...)
	return len(p), nil
}

func (w *responseWriter) WriteHeader(statusCode int) {
	if w.headerWritten {
		return
	}
	w.statusCode = statusCode
	w.headerWritten = true
}

func (w *responseWriter) WriteHeaderIfNotWritten() {
	if !w.headerWritten {
		w.statusCode = http.StatusOK
		w.headerWritten = true
	}
}

func (w *responseWriter) WriteResponse() {
	if !w.headerWritten {
		w.WriteHeader(http.StatusOK)
	}

	statusText := http.StatusText(w.statusCode)
	w.conn.Write([]byte(fmt.Sprintf("HTTP/1.1 %d %s\r\n", w.statusCode, statusText)))

	if _, ok := w.headers["Content-Type"]; !ok {
		w.headers.Set("Content-Type", "text/html; charset=utf-8")
	}
	if _, ok := w.headers["Content-Length"]; !ok {
		w.headers.Set("Content-Length", fmt.Sprintf("%d", len(w.body)))
	}

	for key, values := range w.headers {
		for _, value := range values {
			w.conn.Write([]byte(fmt.Sprintf("%s: %s\r\n", key, value)))
		}
	}

	w.conn.Write([]byte("\r\n"))
	w.conn.Write(w.body)
}

func (pd *ProtocolDetector) handleHTTPRedirect(firstByte []byte, conn net.Conn) {
	reader := bufio.NewReader(&prefixedConn{prefix: firstByte, conn: conn})
	line, err := reader.ReadString('\n')
	if err != nil {
		return
	}
	parts := strings.SplitN(strings.TrimSpace(line), " ", 3)
	if len(parts) < 2 {
		return
	}
	path := parts[1]
	host := ""
	for {
		headerLine, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		headerLine = strings.TrimSpace(headerLine)
		if headerLine == "" {
			break
		}
		if strings.HasPrefix(strings.ToLower(headerLine), "host:") {
			host = strings.TrimSpace(strings.SplitN(headerLine, ":", 2)[1])
			break
		}
	}
	if host == "" {
		host = "localhost"
	}
	if !strings.Contains(host, ":") {
		_, port, _ := net.SplitHostPort(conn.LocalAddr().String())
		host = fmt.Sprintf("%s:%s", host, port)
	}
	target := fmt.Sprintf("https://%s%s", host, path)
	w := bufio.NewWriter(conn)
	_, _ = fmt.Fprintf(w, "HTTP/1.1 307 Temporary Redirect\r\n")
	_, _ = fmt.Fprintf(w, "Location: %s\r\n", target)
	_, _ = fmt.Fprintf(w, "Connection: close\r\n")
	_, _ = fmt.Fprintf(w, "Content-Length: 0\r\n")
	_, _ = fmt.Fprintf(w, "Timing-Allow-Origin: *\r\n")
	_, _ = fmt.Fprintf(w, "Non-Authoritative-Reason: HSTS\r\n")
	_, _ = fmt.Fprintf(w, "x-xss-protection: 0\r\n")
	_, _ = fmt.Fprintf(w, "Cross-Origin-Resource-Policy: Cross-Origin\r\n")
	_, _ = fmt.Fprintf(w, "\r\n")
	w.Flush()
}
