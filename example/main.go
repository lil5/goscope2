package main

import (
	"github.com/gin-gonic/gin"
	"github.com/lil5/goscope2"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	db, _ := gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})
	r := gin.New()
	goscope2.New(goscope2.Config{
		DB: db,
		AllowedApps: map[int32][]string{
			104365: {"localhost:8080"},
		},
		InternalApp: 104365,
		LimitLogs:   3000,
		AuthUser:    "admin",
		AuthPass:    "admin",
	})

	goscope2.AddRoutes(&r.RouterGroup)

	// goscope2.Infof("Run info")
	// goscope2.Warningf("Run warning")
	// goscope2.Errorf("Run error")
	// goscope2.Fatalf("Run fatal")

	r.Run("localhost:8080")
}
