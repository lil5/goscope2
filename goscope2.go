package goscope2

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"text/template"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	"gorm.io/gorm"
)

type Config struct {
	DB *gorm.DB
	// default is 700
	LimitLogs int
	// Create a dictionary of the app id as key and app locations
	// e.g.:
	// ```
	// map[int][]string{239401: []string{"google.com"}}
	// ```
	AllowedApps map[int32][]string
	InternalApp int32
	AuthUser    string
	AuthPass    string
}

var Goscope2 struct{ config *Config }

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
		App:       app,
		Hash:      generateMessageHash(body.Message),
		Severity:  body.Severity,
		Message:   body.Message,
		URL:       c.Request.Host,
		Origin:    c.Request.RemoteAddr,
		UserAgent: c.Request.Header.Get("User-Agent"),
	})

	checkAndPurge(r.DB, r.LimitLogs)
}

var (
	//go:embed admin.html logo.webp tailwind.min.css
	res      embed.FS
	pages, _ = template.ParseFS(res, "*")
)

func (r *Routes) Admin(c *gin.Context) {
	user, pass, ok := c.Request.BasicAuth()
	if !(ok && user == r.AuthUser && pass == r.AuthPass) {
		c.Writer.Header().Set("WWW-Authenticate", "Basic realm=Restricted")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	list, err := getAll(r.DB)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	buf := new(bytes.Buffer)
	jsonAllowedApps, err1 := json.Marshal(r.Config.AllowedApps)
	jsonList, err2 := json.Marshal(list)
	if err1 != nil || err2 != nil {
		glog.Fatalf("Unable to stringify to json %v %v", err1, err2)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	err = pages.ExecuteTemplate(buf, "admin.html", map[string]any{
		"AllowedApps": string(jsonAllowedApps),
		"List":        string(jsonList),
	})
	if err != nil {
		glog.Fatalf("Unable to find template %v", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.DataFromReader(http.StatusOK, int64(buf.Len()), "text/html", buf, nil)
}

func (r *Routes) Favicon(c *gin.Context) {
	c.FileFromFS("logo.webp", http.FS(res))
}
func (r *Routes) Tailwind(c *gin.Context) {
	c.FileFromFS("tailwind.min.css", http.FS(res))
}
