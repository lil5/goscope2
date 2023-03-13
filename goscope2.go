package goscope2

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lil5/goscope2/frontend"
	ginmiddleware "github.com/lil5/goscope2/gin-middleware"
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
	g.GET("/goscope2/index.css", r.Tailwind)
	g.GET("/goscope2/index.js", r.AdminJs)

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
	assetHtml = frontend.MustAsset("dist/index.html")
	assetLogo = frontend.MustAsset("dist/logo.webp")
	assetCss  = frontend.MustAsset("dist/index.css")
	assetJs   = frontend.MustAsset("dist/index.js")
)

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

	c.Data(http.StatusOK, "text/html", assetHtml)
}
func (r *routes) AdminJs(c *gin.Context) {
	c.Data(http.StatusOK, "text/javascript", assetJs)
}

func (r *routes) Favicon(c *gin.Context) {
	c.Data(http.StatusOK, "image/webp", assetLogo)
}
func (r *routes) Tailwind(c *gin.Context) {
	c.Data(http.StatusOK, "text/css", assetCss)
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
