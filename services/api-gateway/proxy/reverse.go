package proxy

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
)

// ReverseProxy creates a reverse proxy handler for a target URL
func ReverseProxy(targetURL string) gin.HandlerFunc {
	target, err := url.Parse(targetURL)
	if err != nil {
		log.Fatalf("Invalid target URL: %s", targetURL)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	// Modify response to remove CORS headers from backend
	proxy.ModifyResponse = func(resp *http.Response) error {
		// Remove CORS headers that backend might have set
		resp.Header.Del("Access-Control-Allow-Origin")
		resp.Header.Del("Access-Control-Allow-Credentials")
		resp.Header.Del("Access-Control-Allow-Methods")
		resp.Header.Del("Access-Control-Allow-Headers")
		resp.Header.Del("Access-Control-Expose-Headers")
		resp.Header.Del("Access-Control-Max-Age")
		return nil
	}

	// Custom error handler
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("Proxy error: %v", err)
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("Service unavailable"))
	}

	return func(c *gin.Context) {
		// Preserve original request path
		c.Request.URL.Host = target.Host
		c.Request.URL.Scheme = target.Scheme
		c.Request.Header.Set("X-Forwarded-Host", c.Request.Header.Get("Host"))
		c.Request.Host = target.Host

		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

// StripPrefix removes a prefix from the request path before proxying
func StripPrefix(prefix string, proxy gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Remove prefix from path
		original := c.Request.URL.Path
		c.Request.URL.Path = original[len(prefix):]

		// Call the proxy
		proxy(c)

		// Restore original path
		c.Request.URL.Path = original
	}
}
