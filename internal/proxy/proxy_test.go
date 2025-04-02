package proxy

import (
	"strings"
	"testing"

	"github.com/Vyary/otel-rev-proxy/internal/models"
)

func TestNewProxy(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		config := &models.Config{
			Routes: map[string]models.Route{
				"example.com": {
					URL:          "http://target.example.com",
					AllowedPaths: []string{"/api/*"},
					Otel:         false,
				},
			},
		}

		proxy, err := NewProxy(config)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		if proxy == nil {
			t.Fatal("Expected proxy object to be created, got nil")
		}
		if len(proxy.proxies) != 1 {
			t.Errorf("Expected 1 proxy, got %d", len(proxy.proxies))
		}
		if _, ok := proxy.proxies["example.com"]; !ok {
			t.Error("Expected 'example.com' to be in proxies map")
		}
	})

	t.Run("invalid URL", func(t *testing.T) {
		config := &models.Config{
			Routes: map[string]models.Route{
				"example.com": {
					URL:          "://invalid-url",
					AllowedPaths: []string{"/api/*"},
					Otel:         false,
				},
			},
		}

		proxy, err := NewProxy(config)

		if err == nil {
			t.Fatal("Expected error for invalid URL, got nil")
		}
		if proxy != nil {
			t.Error("Expected nil proxy when error occurs")
		}
		if !strings.Contains(err.Error(), "Ivalid URL") {
			t.Errorf("Expected error message to contain 'Ivalid URL', got: %s", err.Error())
		}
	})
}

func TestIsPathAllowed(t *testing.T) {
	p := &proxyServer{}

	testCases := []struct {
		name         string
		path         string
		allowedPaths []string
		expected     bool
	}{
		{
			name:         "exact match",
			path:         "/api/v1/users",
			allowedPaths: []string{"/api/v1/users"},
			expected:     true,
		},
		{
			name:         "wildcard all",
			path:         "/any/path",
			allowedPaths: []string{"/*"},
			expected:     true,
		},
		{
			name:         "prefix wildcard match",
			path:         "/api/v1/products",
			allowedPaths: []string{"/api/*"},
			expected:     true,
		},
		{
			name:         "prefix wildcard no match",
			path:         "/auth/login",
			allowedPaths: []string{"/api/*"},
			expected:     false,
		},
		{
			name:         "no match",
			path:         "/private/data",
			allowedPaths: []string{"/public/data", "/api/v1/users"},
			expected:     false,
		},
		{
			name:         "multiple paths with match",
			path:         "/auth/login",
			allowedPaths: []string{"/api/*", "/auth/login", "/public/*"},
			expected:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := p.isPathAllowed(tc.path, tc.allowedPaths)
			if result != tc.expected {
				t.Errorf(
					"Expected isPathAllowed to return %v for path %s with allowed paths %v, got %v",
					tc.expected,
					tc.path,
					tc.allowedPaths,
					result,
				)
			}
		})
	}
}
