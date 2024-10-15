package main

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
	TlsBadRequestLocation string
	TlsBadRequest         string
)

type wrapConn struct {
	net.Conn
}

func ServeTLS(srv *http.Server, l net.Listener, certFile, keyFile string) error {
	config := &tls.Config{}
	if !slices.Contains(config.NextProtos, "http/1.1") {
		config.NextProtos = append(config.NextProtos, "http/1.1")
	}

	configHasCert := len(config.Certificates) > 0 || config.GetCertificate != nil || config.GetConfigForClient != nil
	if !configHasCert || certFile != "" || keyFile != "" {
		var err error
		config.Certificates = make([]tls.Certificate, 1)
		config.Certificates[0], err = tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return err
		}
	}

	tlsListener := tls.NewListener(l, config)
	return srv.Serve(&myListener{Listener: tlsListener})
}

type myListener struct {
	net.Listener
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
