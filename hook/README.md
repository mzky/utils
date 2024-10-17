


// 例子

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
	//srv.SetKeepAlivesEnabled(false)
	//body := "---- message ----"
	//srv.SetResponse(body, func(r *http.Response) {
	//	r.StatusCode = 400
	//	r.Status = http.StatusText(400)
	//	r.Header.Set("Content-Type", "text/html")
	//})

	fmt.Println(srv.ListenAndServeTLS("server.pem", "server.key"))
}
```
