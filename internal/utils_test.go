package internal

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetHeaderApiKey(t *testing.T) {
	tests := []struct {
		name        string
		authHeader  string
		expectedKey string
		expectedErr error
	}{
		{
			name:        "Valid Authorization header",
			authHeader:  "Bearer abc123",
			expectedKey: "abc123",
			expectedErr: nil,
		},
		{
			name:        "Missing Authorization header",
			authHeader:  "",
			expectedKey: "",
			expectedErr: errors.New("missing Authorization header"),
		},
		{
			name:        "Malformed Authorization header",
			authHeader:  "Bearer",
			expectedKey: "",
			expectedErr: errors.New("missing Authorization header"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			key, err := GetHeaderApiKey(nil, req)
			assert.Equal(t, tt.expectedKey, key)
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}
