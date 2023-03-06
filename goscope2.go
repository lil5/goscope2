package goscope2

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"text/template"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	ginmiddleware "github.com/lil5/goscope2/pkg/gin-middleware"
	"gorm.io/gorm"
)

type GoScope2 struct {
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

// Initiates the goscope2 table
// Set defaults if necessary.
func New(gs GoScope2) *GoScope2 {
	// set defaults to config
	if gs.LimitLogs == 0 {
		gs.LimitLogs = 700
	}

	gs.DB.AutoMigrate(&Goscope2Log{})

	return &gs
}

func (gs *GoScope2) AddAdminRoutes(g *gin.RouterGroup) {
	r := &routes{gs}
	g.GET("/goscope2/admin", r.Admin)
	g.GET("/goscope2/logo.webp", r.Favicon)
	g.GET("/goscope2/tailwind.min.css", r.Tailwind)
}

func (gs *GoScope2) AddJsRoute(g *gin.RouterGroup) {
	r := &routes{gs}
	g.POST("/goscope2/", r.PostJsLog)
}

func (gs *GoScope2) AddGinMiddleware(minimumStatus int) func(*gin.Context) {
	return func(c *gin.Context) {
		details := ginmiddleware.ObtainBodyLogWriter(c)
		c.Next()

		status := details.Blw.Status()
		requestPath := c.FullPath()
		if requestPath == "" {
			// Use URL as fallback when path is not recognized as route
			requestPath = c.Request.URL.String()
		}
		if status < minimumStatus || strings.HasPrefix(requestPath, "/goscope") {
			return
		}

		var severity string
		if status >= 500 {
			severity = SEVERITY_ERROR
		} else if status >= 400 && status < 500 {
			severity = SEVERITY_WARNING
		} else {
			severity = SEVERITY_INFO
		}
		log := &Goscope2Log{
			App:       gs.InternalApp,
			Severity:  severity,
			Message:   requestPath,
			Hash:      "",
			URL:       requestPath,
			Origin:    c.ClientIP(),
			UserAgent: c.Request.Header.Get("User-Agent"),
			Status:    status,
		}

		go func() {
			maybeCheckAndPurge(gs.DB, gs.LimitLogs)
			gs.DB.Create(log)
		}()
	}
}

type routes struct{ *GoScope2 }

func (r routes) jsAuth(c *gin.Context) (app int32, ok bool) {
	user, pass, ok := c.Request.BasicAuth()
	if !(ok && user == r.AuthUser && pass == r.AuthPass) {
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

func (r *routes) PostJsLog(c *gin.Context) {
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

	maybeCheckAndPurge(r.DB, r.LimitLogs)
	r.DB.Create(&Goscope2Log{
		App:       app,
		Hash:      generateMessageHash(body.Message),
		Severity:  body.Severity,
		Message:   body.Message,
		URL:       c.Request.Host,
		Origin:    c.Request.RemoteAddr,
		UserAgent: c.Request.Header.Get("User-Agent"),
	})
}

var (
	//go:embed admin.html logo.webp tailwind.min.css
	res      embed.FS
	pages, _ = template.ParseFS(res, "*")
)

func (r *routes) Admin(c *gin.Context) {
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
	jsonAllowedApps, err1 := json.Marshal(r.AllowedApps)
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

func (r *routes) Favicon(c *gin.Context) {
	c.FileFromFS("logo.webp", http.FS(res))
}
func (r *routes) Tailwind(c *gin.Context) {
	c.FileFromFS("tailwind.min.css", http.FS(res))
}
