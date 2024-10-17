package hook

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
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
	defaultBadRequest = "HTTP/1.1 302 Found\r\n\r\n<script>location.protocol='https:'</script>\r\n"
	respBody          string
	redirectPath      = "/"
	resp              = &http.Response{Proto: "HTTP/1.1", Header: make(http.Header)}
)

type handshakeFn func(context.Context) error // (*Conn).clientHandshake or serverHandshake

type listener struct {
	net.Listener
}

type wrapConn struct {
	net.Conn
	response string
}

type Server struct {
	http.Server
}

func NewServer(addr string, handler http.Handler) *Server {
	return &Server{Server: http.Server{Addr: addr, Handler: handler, TLSConfig: &tls.Config{}}}
}

func (srv *Server) ListenAndServeTLS(certFile, keyFile string) error {
	ln, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		return err
	}
	tlsListener := tls.NewListener(ln, srv.TLSConfig)

	configHasCert := len(srv.TLSConfig.Certificates) > 0 || srv.TLSConfig.GetCertificate != nil || srv.TLSConfig.GetConfigForClient != nil
	if !configHasCert || certFile != "" || keyFile != "" {
		var e error
		srv.TLSConfig.Certificates = make([]tls.Certificate, 1)
		srv.TLSConfig.Certificates[0], e = tls.LoadX509KeyPair(certFile, keyFile)
		if e != nil {
			return e
		}
	}
	return srv.Serve(&listener{Listener: tlsListener})
}
func (srv *Server) SetResponse(body string, fn func(r *http.Response)) {
	respBody = body
	fn(resp)
}

func (srv *Server) SetRedirectPath(rp string) {
	redirectPath = rp
}

func (srv *Server) SetDefaultBadRequest(dbr string) {
	defaultBadRequest = dbr
}

// Accept 接受连接并劫持 handshakeFn
func (m *listener) Accept() (net.Conn, error) {
	conn, err := m.Listener.Accept()
	if err != nil {
		return nil, err
	}

	if tl, ok := conn.(*tls.Conn); ok {
		v := reflect.ValueOf(tl).Elem()                               // 获取指向结构体的 Value
		field := v.FieldByName("handshakeFn")                         // 使用反射获取字段
		realPtr := (*handshakeFn)(unsafe.Pointer(field.UnsafeAddr())) // 获取字段的地址并转换为字段类型的指针
		*realPtr = interrupt(*realPtr)                                // 替换 handshakeFn
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
	if c.response == "" {
		c.response = c.response2String()
	}
	return io.WriteString(c.Conn, c.response)
}

// Response2Bytes 构造 HTTP 响应字符串
func (c *wrapConn) response2String() string {
	if respBody == "" {
		return defaultBadRequest
	}
	u := url.URL{Scheme: "https", Host: c.LocalAddr().String(), Path: redirectPath}
	if resp.StatusCode == 0 {
		resp.Status = http.StatusText(302)
		resp.StatusCode = 302
	}
	resp.Header.Set("Location", u.String())
	resp.Header.Set("Date", time.Now().UTC().Format(http.TimeFormat))
	rb := strings.ReplaceAll(respBody, "%s", u.String())
	resp.Header.Set("Content-Length", strconv.Itoa(len(rb)))

	var buffer bytes.Buffer
	fmt.Fprintf(&buffer, "%s %d %s\r\n", resp.Proto, resp.StatusCode, resp.Status)
	for k, vv := range resp.Header {
		for _, v := range vv {
			fmt.Fprintf(&buffer, "%s: %s\r\n", k, v)
		}
	}
	fmt.Fprint(&buffer, "\r\n", respBody, "\r\n")
	return buffer.String()
}
