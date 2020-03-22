package ctxlogger

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/thanhhh/gin-requestid"
)

func AddZerologLoggerToRequestContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		r := c.Request
		ctx := r.Context()

		requestId := requestid.GetReqID(c)

		sublogger := log.Logger.With().
			Str("request-id", requestId).
			Logger()
		newCtx := sublogger.WithContext(ctx)

		c.Request = r.WithContext(newCtx)

		c.Next()
	}
}