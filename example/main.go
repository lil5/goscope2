package main

import (
	"flag"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lil5/goscope2"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	flag.Parse()
	db, _ := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	r := gin.New()
	gs := goscope2.New(goscope2.GoScope2{
		DB:        db,
		JsToken:   "104365",
		LimitLogs: 300,
		AuthUser:  "admin",
		AuthPass:  "admin",
	})

	gs.AddRoutes(&r.RouterGroup)
	r.Use(gs.AddGinMiddleware(http.StatusOK))

	r.GET("/", func(c *gin.Context) {
		gs.Infof("Run info")
		gs.Warningf("Run warning")
		gs.Errorf("Run error")
		// goscope2.Fatalf("Run fatal")

		c.Data(http.StatusOK, "text/html", []byte(`<!DOCTYPE html>
<html>
<head></head>
<body>
<button id="test">Click me</button>
<script>
document.getElementById("test").onclick = function(){
	fetch('/goscope2/js',{
		method: 'post',
		body: JSON.stringify({
			"severity": "WARNING",
			"message": "This is a test from javascript",
			"token": "104365",
		}),
	});
};
</script>
</body>
</html>`))
	})

	r.Run("localhost:8080")
}
