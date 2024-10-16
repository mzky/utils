package hook

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

var (
	DefaultBadRequest = "HTTP/1.1 302 Found\r\n\r\n<script>location.protocol='https:'</script>\r\n"
	Code              int
	HtmlBody          string
	RedirectPath      = "/"
	resp              = &http.Response{Proto: "HTTP/1.1", Header: make(http.Header)}
)

type handshakeFn func(context.Context) error // (*Conn).clientHandshake or serverHandshake

type myListener struct {
	net.Listener
}

type wrapConn struct {
	net.Conn
	response []byte
}

type Server struct {
	http.Server
}

func (srv *Server) ListenAndServeTLS(certFile, keyFile string) error {
	ln, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		return err
	}
	tlsListener := tls.NewListener(ln, srv.TLSConfig)

	if srv.TLSConfig == nil {
		srv.TLSConfig = &tls.Config{}
	}

	configHasCert := len(srv.TLSConfig.Certificates) > 0 || srv.TLSConfig.GetCertificate != nil || srv.TLSConfig.GetConfigForClient != nil
	if !configHasCert || certFile != "" || keyFile != "" {
		var e error
		srv.TLSConfig.Certificates = make([]tls.Certificate, 1)
		srv.TLSConfig.Certificates[0], e = tls.LoadX509KeyPair(certFile, keyFile)
		if e != nil {
			return e
		}
	}
	return srv.Serve(&myListener{Listener: tlsListener})
}
func SetResponse(fn func(r *http.Response)) {
	fn(resp)
}

// Accept 接受连接并劫持 handshakeFn
func (m *myListener) Accept() (net.Conn, error) {
	conn, err := m.Listener.Accept()
	if err != nil {
		return nil, err
	}

	if tl, ok := conn.(*tls.Conn); ok {
		v := reflect.ValueOf(tl).Elem()       // 获取指向结构体的 Value
		field := v.FieldByName("handshakeFn") // 使用反射获取字段
		if !field.IsValid() || field.Kind() != reflect.Func {
			return nil, fmt.Errorf("handshakeFn field not found or invalid")
		}
		ptr := unsafe.Pointer(field.UnsafeAddr()) // 获取字段的地址
		realPtr := (*handshakeFn)(ptr)            // 转换为字段类型的指针
		*realPtr = interrupt(*realPtr)            // 替换 handshakeFn
	}

	return conn, nil
}

func interrupt(fn handshakeFn) handshakeFn {
	return func(ctx context.Context) error {
		err := fn(ctx)
		var re tls.RecordHeaderError
		if errors.As(err, &re) && re.Conn != nil {
			re.Conn = &wrapConn{
				Conn: re.Conn,
			}
			err = re
		}
		return err
	}
}

func (c *wrapConn) Write([]byte) (int, error) {
	if c.response == nil {
		c.response = c.response2Bytes()
	}
	return c.Conn.Write(c.response)
}

// Response2Bytes 构造 HTTP 响应字符串
func (c *wrapConn) response2Bytes() []byte {
	if HtmlBody == "" {
		return []byte(DefaultBadRequest)
	}
	u := url.URL{Scheme: "https", Host: c.LocalAddr().String(), Path: RedirectPath}
	code := IsEmpty(Code, 302)
	resp.Status = http.StatusText(code)
	resp.StatusCode = code
	resp.Header.Set("Location", u.String())
	resp.Header.Set("Date", time.Now().UTC().Format(http.TimeFormat))
	htmlBody := strings.ReplaceAll(HtmlBody, "%s", u.String())
	resp.Header.Set("Content-Length", strconv.Itoa(len(htmlBody)))

	var buffer bytes.Buffer
	fmt.Fprintf(&buffer, "%s %d %s\r\n", resp.Proto, resp.StatusCode, resp.Status)
	for k, vv := range resp.Header {
		for _, v := range vv {
			fmt.Fprintf(&buffer, "%s: %s\r\n", k, v)
		}
	}
	fmt.Fprintf(&buffer, "\r\n")
	fmt.Fprintf(&buffer, htmlBody)
	fmt.Fprintf(&buffer, "\r\n")
	return buffer.Bytes()
}

func IsEmpty(value int, def int) int {
	if value == 0 {
		return def
	}
	return value
}
func IsEmptyString(value string, def string) string {
	if value == "" {
		return def
	}
	return value
}
