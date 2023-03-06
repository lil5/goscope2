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
	db, _ := gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})
	r := gin.New()
	gs := goscope2.New(goscope2.GoScope2{
		DB: db,
		AllowedApps: map[int32][]string{
			104365: {"localhost:8080"},
		},
		InternalApp: 104365,
		LimitLogs:   3000,
		AuthUser:    "admin",
		AuthPass:    "admin",
	})

	gs.AddAdminRoutes(&r.RouterGroup)
	gs.AddJsRoute(&r.RouterGroup)
	r.Use(gs.AddGinMiddleware(http.StatusOK))

	r.GET("/", func(ctx *gin.Context) {
		gs.Infof("Run info")
		gs.Warningf("Run warning")
		gs.Errorf("Run error")
		// goscope2.Fatalf("Run fatal")
	})

	r.Run("localhost:8080")
}
