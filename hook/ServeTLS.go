package hook

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"slices"
	"unsafe"
)

var (
	TlsBadRequestLocation string // 重定向的路径
	TlsBadRequest         string // 浏览器返回的错误信息
)

type wrapConn struct {
	net.Conn
}

type myListener struct {
	net.Listener
}

type Server struct {
	http.Server
}

func (srv *Server) ServeTLS(certFile, keyFile string) error {
	addr := srv.Addr
	if addr == "" {
		addr = ":http"
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	tlsListener := tls.NewListener(ln, srv.TLSConfig)

	if srv.TLSConfig == nil {
		srv.TLSConfig = &tls.Config{}
	}
	if !slices.Contains(srv.TLSConfig.NextProtos, "http/1.1") {
		srv.TLSConfig.NextProtos = append(srv.TLSConfig.NextProtos, "http/1.1")
	}

	configHasCert := len(srv.TLSConfig.Certificates) > 0 || srv.TLSConfig.GetCertificate != nil || srv.TLSConfig.GetConfigForClient != nil
	if !configHasCert || certFile != "" || keyFile != "" {
		var err error
		srv.TLSConfig.Certificates = make([]tls.Certificate, 1)
		srv.TLSConfig.Certificates[0], err = tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return err
		}
	}
	return srv.Serve(&myListener{Listener: tlsListener})
}

func (m *myListener) Accept() (net.Conn, error) {
	conn, err := m.Listener.Accept()
	if tl, ok := conn.(*tls.Conn); ok {
		v := reflect.ValueOf(tl).Elem()           // 获取指向结构体的 Value
		field := v.FieldByName("handshakeFn")     // 使用 unsafe 绕过访问控制
		ptr := unsafe.Pointer(field.UnsafeAddr()) // 获取字段的地址
		realPtr := (*handshakeFn)(ptr)            // 转换为字段类型的指针
		*realPtr = interrupt(*realPtr)
	}
	return conn, err
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

func (c *wrapConn) Write([]byte) (n int, err error) {
	u := url.URL{Scheme: "https", Host: c.Conn.LocalAddr().String(), Path: TlsBadRequestLocation}
	if TlsBadRequest == "" {
		TlsBadRequest = fmt.Sprintf(`HTTP/1.1 307 Internal Redirect
Location: %s
Timing-Allow-Origin: *
Cross-Origin-Resource-Policy: Cross-Origin
Non-Authoritative-Reason: HSTS
x-xss-protection: 0

<head><meta http-equiv="refresh" content="1;url="%s"></head>`, u.String(), u.String())
	}
	return io.WriteString(c.Conn, TlsBadRequest)
}

type handshakeFn func(context.Context) error // (*Conn).clientHandshake or serverHandshake
