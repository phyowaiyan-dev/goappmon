package utils

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestJSONHelpers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("json", func(t *testing.T) {
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		JSON(c, http.StatusCreated, gin.H{"status": "ok"})
		if rec.Code != http.StatusCreated {
			t.Fatalf("unexpected code: %d", rec.Code)
		}
		if rec.Body.String() == "" {
			t.Fatal("expected body")
		}
	})

	t.Run("json error", func(t *testing.T) {
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		JSONError(c, http.StatusBadRequest, "bad request")
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("unexpected code: %d", rec.Code)
		}
	})
}
