package goscope2

import (
	"github.com/gin-gonic/gin"
)

// Initiates the goscope2 table
// Set defaults if necessary.
func New(c Config) {
	// set defaults to config
	if c.LimitLogs == 0 {
		c.LimitLogs = 700
	}

	c.DB.AutoMigrate(&Goscope2Log{})

	Goscope2.config = &c
}

func AddRoutes(r *gin.RouterGroup) {
	routes := &Routes{Goscope2.config}
	r.POST("/goscope2/", routes.PostJsLog)
	r.GET("/goscope2/admin", routes.Admin)
	r.GET("/goscope2/logo.webp", routes.Favicon)
	r.GET("/goscope2/tailwind.min.css", routes.Tailwind)
}
