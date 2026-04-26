package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

var limiter = rate.NewLimiter(rate.Limit(50), 200)

func RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 对 OPTIONS 请求直接放行，不进行限流
		if c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}
		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "请求过于频繁,请稍后再试",
			})
			return
		}
		c.Next()
	}
}
