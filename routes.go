package goscope2

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Routes struct{ *Config }

func (r Routes) jsAuth(c *gin.Context) (app int32, ok bool) {
	user, pass, ok := c.Request.BasicAuth()
	if !ok || user != "goscope2" {
		c.AbortWithStatus(http.StatusUnauthorized)
		return 0, false
	}
	fmt.Println(c.Request.Host)
	for id, addrs := range r.AllowedApps {
		if strconv.Itoa(int(id)) == pass {
			for _, addr := range addrs {
				if addr == c.Request.Host {
					return id, true
				}
			}
		}
	}

	c.AbortWithStatus(http.StatusUnauthorized)
	return 0, false
}

func (r *Routes) PostJsLog(c *gin.Context) {
	app, ok := r.jsAuth(c)
	if !ok {
		return
	}

	var body struct {
		Severity string `json:"severity" binding:"required,oneof=INFO WARNING ERROR FATAL"`
		Message  string `json:"message" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.String(http.StatusBadRequest, fmt.Sprint(err))
		return
	}

	r.DB.Create(&Goscope2Log{
		App:         app,
		MessageHash: generateMessageHash(body.Message),
		Severity:    body.Severity,
		Message:     body.Message,
		URL:         c.Request.Host,
		Origin:      c.Request.RemoteAddr,
		UserAgent:   c.Request.Header.Get("User-Agent"),
	})

	checkAndPurge(r.DB, r.LimitLogs)
}
