package utils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSendHelpers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/success", func(c *gin.Context) { SendSuccess(c, gin.H{"x": 1}, "ok") })
	r.GET("/created", func(c *gin.Context) { SendCreated(c, gin.H{"x": 2}, "created") })
	r.GET("/bad", func(c *gin.Context) { SendBadRequest(c, "bad") })
	r.GET("/notfound", func(c *gin.Context) { SendNotFound(c, "nf") })
	r.GET("/conflict", func(c *gin.Context) { SendConflict(c, "cf") })
	r.GET("/internal", func(c *gin.Context) { SendInternalError(c, "ie") })

	cases := []struct {
		path string
		code int
	}{
		{"/success", http.StatusOK},
		{"/created", http.StatusCreated},
		{"/bad", http.StatusBadRequest},
		{"/notfound", http.StatusNotFound},
		{"/conflict", http.StatusConflict},
		{"/internal", http.StatusInternalServerError},
	}
	for _, cs := range cases {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, cs.path, nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, cs.code, w.Code)
		var m map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &m)
		assert.Contains(t, m, "success")
	}
}
