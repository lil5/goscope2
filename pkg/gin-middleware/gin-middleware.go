package ginmiddleware

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"

	"github.com/gin-gonic/gin"
)

type BodyLogWriter struct {
	gin.ResponseWriter
	Body *bytes.Buffer
}

type BodyLogWriterResponse struct {
	Blw *BodyLogWriter
	Rdr io.ReadCloser
}

func ObtainBodyLogWriter(c *gin.Context) BodyLogWriterResponse {
	blw := &BodyLogWriter{Body: bytes.NewBufferString(""), ResponseWriter: c.Writer}

	c.Writer = blw

	buf, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		log.Println(err.Error())
	}

	rdr1 := ioutil.NopCloser(bytes.NewBuffer(buf))
	// We have to create a new Buffer, because rdr1 will be read and consumed.
	rdr2 := ioutil.NopCloser(bytes.NewBuffer(buf))
	c.Request.Body = rdr2

	return BodyLogWriterResponse{
		Blw: blw,
		Rdr: rdr1,
	}
}
func ReadBody(reader io.Reader) string {
	buf := new(bytes.Buffer)

	_, err := buf.ReadFrom(reader)
	if err != nil {
		log.Println(err.Error())
	}

	s := buf.String()

	return s
}
