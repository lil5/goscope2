package goscope2

import (
	"bytes"
	"embed"
	"encoding/json"
	"net/http"
	"text/template"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
)

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
