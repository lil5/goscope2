# GoScope2

The second iteration from where josep left off

A log tracker, with a basic ui.

Requireds golangs **gin** http framework and web **api calls**.

## Basic Usage

```golang
package main

import (
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
		DB:            db,
		JsToken:       "104365",
		LimitLogs:     3000,
		AuthUser:      "admin",
		AuthPass:      "admin",
	})

	gs.AddRoutes(&r.RouterGroup)
	r.Use(gs.AddGinMiddleware(http.StatusOK))

	gs.Infof("Run info")
	gs.Warningf("Run warning")
	gs.Errorf("Run error")
	// goscope2.Fatalf("Run fatal")

	r.Run("localhost:8080")
}
```

## License

Mozilla Public License 2.0
