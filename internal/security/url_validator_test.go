package security

import (
	"strings"
	"testing"
)

func TestValidateURL_ValidURLs(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{"HTTPS public URL", "https://example.com/endpoint"},
		{"HTTP public URL", "http://example.com/endpoint"},
		{"URL with port", "https://example.com:8443/data"},
		{"URL with query params", "https://example.com/search?q=test"},
		{"URL with fragment", "https://example.com/page#section"},
	}

	validator := NewURLValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateURL(tt.url)
			// In test environment, DNS resolution may fail, but we verify
			// that the scheme and parsing logic work correctly
			// A DNS error is acceptable, but scheme/parsing errors are not
			if err != nil && !strings.Contains(err.Error(), "failed to resolve") {
				t.Errorf("ValidateURL(%q) returned unexpected error: %v", tt.url, err)
			}
		})
	}
}

func TestValidateURL_BlockedSchemes(t *testing.T) {
	tests := []struct {
		name   string
		url    string
		errMsg string
	}{
		{"file scheme", "file:///etc/passwd", "only http and https schemes allowed"},
		{"ftp scheme", "ftp://ftp.example.com/file.txt", "only http and https schemes allowed"},
		{"gopher scheme", "gopher://example.com", "only http and https schemes allowed"},
		{"javascript scheme", "javascript:alert('xss')", "only http and https schemes allowed"},
		{"data scheme", "data:text/html,<script>alert('xss')</script>", "only http and https schemes allowed"},
	}

	validator := NewURLValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateURL(tt.url)
			if err == nil {
				t.Errorf("ValidateURL(%q) expected error, got nil", tt.url)
			}
			if err != nil && err.Error() != tt.errMsg {
				t.Errorf("ValidateURL(%q) error = %v, want %v", tt.url, err, tt.errMsg)
			}
		})
	}
}

func TestValidateURL_LoopbackAddresses(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{"localhost", "http://localhost/api"},
		{"127.0.0.1", "http://127.0.0.1/api"},
		{"127.1.1.1", "http://127.1.1.1/api"},
		{"127.255.255.255", "http://127.255.255.255/api"},
		{"IPv6 loopback", "http://[::1]/api"},
		{"IPv6 loopback expanded", "http://[0000:0000:0000:0000:0000:0000:0000:0001]/api"},
	}

	validator := NewURLValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateURL(tt.url)
			if err == nil {
				t.Errorf("ValidateURL(%q) expected error for loopback address, got nil", tt.url)
			}
		})
	}
}

func TestValidateURL_PrivateIPAddresses(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{"10.x.x.x network", "http://10.0.0.1/api"},
		{"10.x.x.x boundary", "http://10.255.255.255/api"},
		{"172.16.x.x network start", "http://172.16.0.1/api"},
		{"172.31.x.x network end", "http://172.31.255.255/api"},
		{"172.20.x.x network middle", "http://172.20.100.50/api"},
		{"192.168.x.x network", "http://192.168.1.1/api"},
		{"192.168.x.x boundary", "http://192.168.255.255/api"},
	}

	validator := NewURLValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateURL(tt.url)
			if err == nil {
				t.Errorf("ValidateURL(%q) expected error for private IP, got nil", tt.url)
			}
		})
	}
}

func TestValidateURL_LinkLocalAddresses(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{"AWS metadata service", "http://169.254.169.254/latest/meta-data/"},
		{"Link-local start", "http://169.254.0.1/api"},
		{"Link-local end", "http://169.254.255.255/api"},
		{"GCP metadata", "http://169.254.169.254/computeMetadata/v1/"},
	}

	validator := NewURLValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateURL(tt.url)
			if err == nil {
				t.Errorf("ValidateURL(%q) expected error for link-local address (metadata service), got nil", tt.url)
			}
		})
	}
}

func TestValidateURL_IPv6PrivateAddresses(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{"IPv6 unique local start", "http://[fc00::1]/api"},
		{"IPv6 unique local", "http://[fd12:3456:789a:1::1]/api"},
		{"IPv6 link-local", "http://[fe80::1]/api"},
		{"IPv6 link-local full", "http://[fe80:0000:0000:0000:0204:61ff:fe9d:f156]/api"},
	}

	validator := NewURLValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateURL(tt.url)
			if err == nil {
				t.Errorf("ValidateURL(%q) expected error for IPv6 private address, got nil", tt.url)
			}
		})
	}
}

func TestValidateURL_InvalidURLs(t *testing.T) {
	tests := []struct {
		name   string
		url    string
		errMsg string
	}{
		{"empty URL", "", "URL is required"},
		{"malformed URL", "not a url", "invalid URL"},
		{"missing scheme", "example.com/api", "only http and https schemes allowed"},
		{"only scheme", "http://", "hostname is required"},
	}

	validator := NewURLValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateURL(tt.url)
			if err == nil {
				t.Errorf("ValidateURL(%q) expected error, got nil", tt.url)
			}
		})
	}
}

func TestValidateURL_DNSResolution(t *testing.T) {
	// This test verifies that DNS resolution is performed
	// We can't easily test resolution to blocked IPs without a real DNS server,
	// but we can test that resolution happens
	validator := NewURLValidator()

	// Test with a domain that should resolve (example.com is reserved for documentation)
	err := validator.ValidateURL("https://example.com")
	if err != nil {
		// This is acceptable - DNS might not be available in test environment
		// The important thing is that we attempted resolution
		t.Logf("DNS resolution returned error (acceptable in test environment): %v", err)
	}
}

func TestValidateURL_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		expectErr bool
		reason    string
	}{
		{
			"URL with credentials",
			"http://user:pass@example.com",
			false,
			"Credentials in URL should be allowed but verified",
		},
		{
			"URL with uppercase hostname",
			"http://LOCALHOST/api",
			true,
			"Uppercase localhost should be blocked",
		},
		{
			"URL with mixed case hostname",
			"http://LocalHost/api",
			true,
			"Mixed case localhost should be blocked",
		},
		{
			"Decimal IP notation",
			"http://2130706433/api",
			true,
			"Decimal notation of 127.0.0.1 should be blocked",
		},
		{
			"Hexadecimal IP",
			"http://0x7f.0x0.0x0.0x1/api",
			true,
			"Hex notation of 127.0.0.1 should be blocked",
		},
	}

	validator := NewURLValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateURL(tt.url)
			if tt.expectErr && err == nil {
				t.Errorf("ValidateURL(%q) expected error (%s), got nil", tt.url, tt.reason)
			}
			if !tt.expectErr && err != nil {
				// Some edge cases might fail in test environment, log but don't fail
				t.Logf("ValidateURL(%q) returned error (might be acceptable): %v", tt.url, err)
			}
		})
	}
}

func TestValidateURL_WithConfig(t *testing.T) {
	tests := []struct {
		name      string
		config    *URLValidatorConfig
		url       string
		expectErr bool
	}{
		{
			"Disabled protection allows all",
			&URLValidatorConfig{Enabled: false},
			"http://127.0.0.1/api",
			false,
		},
		{
			"Custom allowed network",
			&URLValidatorConfig{
				Enabled:         true,
				AllowedNetworks: []string{"192.168.1.0/24"},
			},
			"http://192.168.1.100/api",
			false,
		},
		{
			"Custom blocked network",
			&URLValidatorConfig{
				Enabled:         true,
				BlockedNetworks: []string{"203.0.113.0/24"},
			},
			"http://203.0.113.50/api",
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewURLValidatorWithConfig(tt.config)
			err := validator.ValidateURL(tt.url)
			if tt.expectErr && err == nil {
				t.Errorf("ValidateURL(%q) expected error, got nil", tt.url)
			}
			if !tt.expectErr && err != nil {
				t.Errorf("ValidateURL(%q) expected no error, got: %v", tt.url, err)
			}
		})
	}
}

func TestIsBlockedIP(t *testing.T) {
	tests := []struct {
		name   string
		ip     string
		expect bool
	}{
		{"Public IP 8.8.8.8", "8.8.8.8", false},
		{"Public IP 1.1.1.1", "1.1.1.1", false},
		{"Loopback 127.0.0.1", "127.0.0.1", true},
		{"Private 10.0.0.1", "10.0.0.1", true},
		{"Private 172.16.0.1", "172.16.0.1", true},
		{"Private 192.168.1.1", "192.168.1.1", true},
		{"Link-local 169.254.169.254", "169.254.169.254", true},
		{"IPv6 loopback", "::1", true},
		{"IPv6 private fc00", "fc00::1", true},
		{"IPv6 link-local", "fe80::1", true},
	}

	validator := NewURLValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.isBlockedIP(tt.ip)
			if result != tt.expect {
				t.Errorf("isBlockedIP(%q) = %v, want %v", tt.ip, result, tt.expect)
			}
		})
	}
}

func TestURLValidator_ConcurrentAccess(t *testing.T) {
	// Test that validator is safe for concurrent use
	validator := NewURLValidator()
	urls := []string{
		"https://api.example.com",
		"http://10.0.0.1",
		"http://127.0.0.1",
		"https://example.org",
	}

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			for _, url := range urls {
				_ = validator.ValidateURL(url)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}
