package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestValidateArgocdHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name            string
		headers         map[string]string
		url             string
		expectedCode    int
		expectedMessage string
	}{
		{
			name: "valid headers and URL",
			headers: map[string]string{
				headerArgocdAppName:       "test-app",
				headerArgocdProjName:      "test-project",
				headerArgocdExtensionName: "test-extension",
			},
			url:             APIPathV1 + apiPatchTouch + "/test-extension/",
			expectedCode:    http.StatusOK,
			expectedMessage: "",
		},
		{
			name: "missing ArgoCD app name header",
			headers: map[string]string{
				headerArgocdProjName:      "test-project",
				headerArgocdExtensionName: "test-extension",
			},
			url:             APIPathV1 + apiPatchTouch + "/test-extension/",
			expectedCode:    http.StatusBadRequest,
			expectedMessage: "",
		},
		{
			name: "missing ArgoCD project name header",
			headers: map[string]string{
				headerArgocdAppName:       "test-app",
				headerArgocdExtensionName: "test-extension",
			},
			url:             APIPathV1 + apiPatchTouch + "/test-extension/",
			expectedCode:    http.StatusBadRequest,
			expectedMessage: "",
		},
		{
			name: "missing ArgoCD extension name header",
			headers: map[string]string{
				headerArgocdAppName:  "test-app",
				headerArgocdProjName: "test-project",
			},
			url:             APIPathV1 + apiPatchTouch + "/test-extension/",
			expectedCode:    http.StatusBadRequest,
			expectedMessage: "",
		},
		{
			name: "invalid URL path for extension",
			headers: map[string]string{
				headerArgocdAppName:       "test-app",
				headerArgocdProjName:      "test-project",
				headerArgocdExtensionName: "test-extension",
			},
			url:             APIPathV1 + apiPatchTouch + "/invalid-extension/",
			expectedCode:    http.StatusBadRequest,
			expectedMessage: "Invalid extension name: test-extension",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(validateArgocdHeaders())

			router.GET(APIPathV1+apiPatchTouch+"/:name/", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, tt.url, http.NoBody)
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedCode, rec.Code)
			if tt.expectedMessage != "" {
				assert.Contains(t, rec.Body.String(), tt.expectedMessage)
			}
		})
	}
}
