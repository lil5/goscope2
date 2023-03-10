package goscope2

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/Masterminds/sprig/v3"
	"github.com/gin-gonic/gin"
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

func (gs *GoScope2) AddRoutes(g *gin.RouterGroup) {
	r := &routes{gs}
	g.GET("/goscope2/", r.Admin)
	g.GET("/goscope2/logo.webp", r.Favicon)
	g.GET("/goscope2/tailwind.min.css", r.Tailwind)

	g.GET("/goscope2/api", r.ApiGet)
	g.POST("/goscope2/api", r.ApiCreate)

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

		log := &Goscope2Log{
			Type:      TYPE_HTTP,
			Message:   requestPath,
			URL:       requestPath,
			Origin:    c.ClientIP(),
			UserAgent: c.Request.Header.Get("User-Agent"),
			Status:    status,
		}

		go func() {
			if status >= 500 {
				log.Severity = SEVERITY_ERROR
			} else if status >= 400 && status < 500 {
				log.Severity = SEVERITY_WARNING
			} else {
				log.Severity = SEVERITY_INFO
			}
			log.GenerateHash()
			gs.DB.Create(log)
		}()
	}
}

type routes struct{ *GoScope2 }

// js route

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

	log := &Goscope2Log{
		Type:      TYPE_JS,
		Severity:  body.Severity,
		Message:   body.Message,
		URL:       c.Request.Host,
		Origin:    c.Request.RemoteAddr,
		UserAgent: c.Request.Header.Get("User-Agent"),
	}
	go func() {
		log.GenerateHash()
		r.DB.Create(log)
	}()
}

// admin routes

var (
	//go:embed admin.html logo.webp tailwind.min.css
	res   embed.FS
	pages *template.Template
)

func init() {
	var err error
	pages, err = template.New("admin").Funcs(sprig.FuncMap()).ParseFS(res, "*")
	if err != nil {
		log.Fatal(err)
	}
}

func (r *routes) adminAuth(c *gin.Context) bool {
	user, pass, ok := c.Request.BasicAuth()
	if !(ok && user == r.AuthUser && pass == r.AuthPass) {
		c.Writer.Header().Set("WWW-Authenticate", "Basic realm=Restricted")
		c.AbortWithStatus(http.StatusUnauthorized)
		return false
	}
	return true
}

func (r *routes) Admin(c *gin.Context) {
	if !r.adminAuth(c) {
		return
	}

	err := checkAndPurge(r.DB, r.LimitLogs)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	buf := new(bytes.Buffer)
	err = pages.ExecuteTemplate(buf, "admin.html", nil)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
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

// api routes

func (r *routes) ApiGet(c *gin.Context) {
	if !r.adminAuth(c) {
		return
	}

	var query struct {
		Page int    `form:"page" binding:"min=1"`
		Type string `form:"type" binding:"required,oneof='http' 'js' 'log'"`
	}
	if err := c.ShouldBindQuery(&query); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	data, err := getSome(r.DB, query.Page, query.Type)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, data)
}

func (r *routes) ApiCreate(c *gin.Context) {
	if !r.adminAuth(c) {
		return
	}

	var body *Goscope2Log
	if err := c.ShouldBindJSON(body); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	body.ID = 0
	body.GenerateHash()

	r.DB.Save(body)
}
