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

type handshakeFn func(context.Context) error // (*Conn).clientHandshake or serverHandshake

type listener struct {
	net.Listener
	server *Server
}

type wrapConn struct {
	net.Conn
	response string
	server   *Server
}

type Server struct {
	http.Server
	defaultBadRequest string
	respBody          string
	redirectPath      string
	resp              *http.Response
}

func NewServer(addr string, handler http.Handler) *Server {
	return &Server{
		Server:            http.Server{Addr: addr, Handler: handler, TLSConfig: &tls.Config{}},
		defaultBadRequest: "HTTP/1.1 302 Found\r\n\r\n<script>location.protocol='https:'</script>\r\n",
		respBody:          "",
		redirectPath:      "/",
		resp:              &http.Response{Proto: "HTTP/1.1", Header: make(http.Header)},
	}
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
	return srv.Serve(&listener{Listener: tlsListener, server: srv})
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
		*realPtr = m.server.interrupt(*realPtr)                       // 替换 handshakeFn
	}

	return conn, nil
}

func (srv *Server) interrupt(fn handshakeFn) handshakeFn {
	return func(ctx context.Context) error {
		err := fn(ctx)
		var re tls.RecordHeaderError
		if errors.As(err, &re) && re.Conn != nil {
			re.Conn = &wrapConn{
				Conn:   re.Conn,
				server: srv,
			}
			err = re
		}
		return err
	}
}

func (c *wrapConn) Write([]byte) (int, error) {
	if c.response == "" {
		c.response = c.response2String(c.server)
	}
	return io.WriteString(c.Conn, c.response)
}

func (srv *Server) SetResponse(body string, fn func(r *http.Response)) {
	srv.respBody = body
	fn(srv.resp)
}

func (srv *Server) SetRedirectPath(rp string) {
	srv.redirectPath = rp
}

func (srv *Server) SetDefaultBadRequest(dbr string) {
	srv.defaultBadRequest = dbr
}

// response2String 构造 HTTP 响应字符串
func (c *wrapConn) response2String(srv *Server) string {
	if srv.respBody == "" && srv.resp.StatusCode == 0 {
		return srv.defaultBadRequest
	}

	u := url.URL{Scheme: "https", Host: c.LocalAddr().String(), Path: srv.redirectPath}
	if srv.resp.StatusCode == 0 {
		srv.resp.Status = http.StatusText(302)
		srv.resp.StatusCode = 302
	}
	if srv.resp.Header.Get("Location") == "" {
		srv.resp.Header.Set("Location", u.String())
	}
	srv.resp.Header.Set("Date", time.Now().UTC().Format(http.TimeFormat))
	rb := strings.ReplaceAll(srv.respBody, "%s", u.String())
	srv.resp.Header.Set("Content-Length", strconv.Itoa(len(rb)))

	var buffer bytes.Buffer
	fmt.Fprintf(&buffer, "%s %d %s\r\n", srv.resp.Proto, srv.resp.StatusCode, srv.resp.Status)
	for k, vv := range srv.resp.Header {
		for _, v := range vv {
			fmt.Fprintf(&buffer, "%s: %s\r\n", k, v)
		}
	}
	fmt.Fprint(&buffer, "\r\n", srv.respBody, "\r\n")
	return buffer.String()
}
