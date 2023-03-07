package goscope2

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"net/http"
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
	LimitLogs     int
	JsToken       string
	AllowedOrigin []string
	AuthUser      string
	AuthPass      string
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
	g.POST("/goscope2/js", r.JsLog)
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
			Type:      TYPE_GIN,
			Severity:  severity,
			Message:   requestPath,
			Hash:      generateMessageHash(requestPath),
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

func (r routes) jsAuth(c *gin.Context) (ok bool) {
	token := c.Request.Header.Get("Token")
	if r.JsToken == "" {
		c.AbortWithStatus(http.StatusNotImplemented)
		return false
	}
	if token != r.JsToken {
		c.AbortWithStatus(http.StatusUnauthorized)
		return false
	}
	fmt.Println(c.Request.Host)
	for _, addr := range r.AllowedOrigin {
		if addr == c.Request.Host {
			return true
		}
	}

	c.AbortWithStatus(http.StatusUnauthorized)
	return false
}

func (r *routes) JsLog(c *gin.Context) {
	if !r.jsAuth(c) {
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
		Type:      TYPE_JS,
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
	jsonList, err := json.Marshal(list)
	if err != nil {
		glog.Fatalf("Unable to stringify to json %v", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	err = pages.ExecuteTemplate(buf, "admin.html", map[string]any{
		"List": string(jsonList),
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
