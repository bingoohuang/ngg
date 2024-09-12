package util

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
)

var staticReg = regexp.MustCompile(".(js|jpg|jpeg|ico|css|woff2|html|woff|ttf|svg|png|eot|map)$") //nolint

// Logger is the logrus logger handler
// Filter static when true
func Logger(filter bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// other handler can change c.Path so:
		path := c.Request.URL.Path
		if filter && staticReg.MatchString(path) {
			c.Next()
			return
		}

		// other handler can change c.Path so:
		start := time.Now()

		c.Next()

		stop := time.Since(start)
		statusCode := c.Writer.Status()

		if len(c.Errors) > 0 {
			log.Printf("E! gin errors: %v", c.Errors.ByType(gin.ErrorTypePrivate).String())
			return
		}

		msg := fmt.Sprintf("%s %s %s [%d] %d %s %s (%s)",
			c.ClientIP(), c.Request.Method, path, statusCode,
			c.Writer.Size(), c.Request.Referer(), c.Request.UserAgent(), stop)

		switch {
		case statusCode > 499:
			log.Printf("E! %s", msg)
		case statusCode > 399:
			log.Printf("W! %s", msg)
		default:
			log.Printf("I! %s", msg)
		}
	}
}
