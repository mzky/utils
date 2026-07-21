package hook

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/tjfoc/gmsm/gmtls"
)

type ProtocolDetector struct {
	TLSConfig *gmtls.Config
	Handler   http.Handler
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
			w.Flush()
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

func (w *responseWriter) Flush() {
	if !w.headerWritten {
		w.WriteHeader(http.StatusOK)
	}

	statusText := http.StatusText(w.statusCode)
	w.conn.Write([]byte(fmt.Sprintf("HTTP/1.1 %d %s\r\n", w.statusCode, statusText)))

	if _, ok := w.headers["Content-Type"]; !ok {
		w.headers.Set("Content-Type", "text/plain; charset=utf-8")
	}
	if _, ok := w.headers["Content-Length"]; !ok {
		w.headers.Set("Content-Length", fmt.Sprintf("%d", len(w.body)))
	}
	if _, ok := w.headers["Connection"]; !ok {
		w.headers.Set("Connection", "close")
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
	fmt.Fprintf(w, "HTTP/1.1 307 Temporary Redirect\r\n")
	fmt.Fprintf(w, "Location: %s\r\n", target)
	fmt.Fprintf(w, "Content-Length: 0\r\n")
	fmt.Fprintf(w, "Timing-Allow-Origin: *\r\n")
	fmt.Fprintf(w, "Non-Authoritative-Reason: HSTS\r\n")
	fmt.Fprintf(w, "x-xss-protection: 0\r\n")
	fmt.Fprintf(w, "Cross-Origin-Resource-Policy: Cross-Origin\r\n")
	fmt.Fprintf(w, "\r\n")

	w.Flush()
}
