package middleware

import "testing"

func TestSanitizePath(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "download token is redacted",
			path: "/api/v1/files/download/abc123secret456",
			want: "/api/v1/files/download/[REDACTED]",
		},
		{
			name: "download token with long value",
			path: "/files/download/eyJhbGciOiJIUzI1NiJ9.very-secret-token",
			want: "/files/download/[REDACTED]",
		},
		{
			name: "normal path not redacted",
			path: "/api/v1/bookings/123",
			want: "/api/v1/bookings/123",
		},
		{
			name: "health path unchanged",
			path: "/health",
			want: "/health",
		},
		{
			name: "files upload path unchanged",
			path: "/api/v1/files/upload",
			want: "/api/v1/files/upload",
		},
		{
			name: "files download prefix but no token",
			path: "/files/download/",
			want: "/files/download/",
		},
		{
			name: "empty path",
			path: "",
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizePath(tt.path)
			if got != tt.want {
				t.Errorf("sanitizePath(%q) = %q, want %q", tt.path, got, tt.want)
			}
			// Token must never appear in sanitized output
			if tt.name == "download token is redacted" {
				if got != tt.want {
					t.Errorf("token leaked in log path: %q", got)
				}
			}
		})
	}
}

func TestMaskQueryParams(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{"empty", "", ""},
		{"normal params", "page=1&pageSize=20", "page=1&pageSize=20"},
		{"token masked", "token=secret123&page=1", "token=***&page=1"},
		{"password masked", "password=hunter2&email=a@b.com", "password=***&email=a@b.com"},
		{"api_key masked", "api_key=mykey&foo=bar", "api_key=***&foo=bar"},
		{"case insensitive", "Token=secret", "Token=***"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := maskQueryParams(tt.raw)
			if got != tt.want {
				t.Errorf("maskQueryParams(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}
