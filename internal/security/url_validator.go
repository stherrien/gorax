package security

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

// URLValidatorConfig holds URL validation configuration
type URLValidatorConfig struct {
	// Enabled controls whether SSRF protection is active
	Enabled bool
	// AllowedNetworks are CIDR ranges that are explicitly allowed (overrides blocklist)
	AllowedNetworks []string
	// BlockedNetworks are additional CIDR ranges to block beyond defaults
	BlockedNetworks []string
}

// URLValidator validates URLs to prevent SSRF attacks
type URLValidator struct {
	config          *URLValidatorConfig
	allowedNetworks []*net.IPNet
	blockedNetworks []*net.IPNet
}

// NewURLValidator creates a new URL validator with default configuration
func NewURLValidator() *URLValidator {
	return NewURLValidatorWithConfig(&URLValidatorConfig{
		Enabled:         true,
		AllowedNetworks: []string{},
		BlockedNetworks: []string{},
	})
}

// NewURLValidatorWithConfig creates a new URL validator with custom configuration
func NewURLValidatorWithConfig(config *URLValidatorConfig) *URLValidator {
	v := &URLValidator{
		config:          config,
		allowedNetworks: make([]*net.IPNet, 0),
		blockedNetworks: make([]*net.IPNet, 0),
	}

	// Parse allowed networks
	for _, cidr := range config.AllowedNetworks {
		_, network, err := net.ParseCIDR(cidr)
		if err == nil {
			v.allowedNetworks = append(v.allowedNetworks, network)
		}
	}

	// Parse blocked networks
	for _, cidr := range config.BlockedNetworks {
		_, network, err := net.ParseCIDR(cidr)
		if err == nil {
			v.blockedNetworks = append(v.blockedNetworks, network)
		}
	}

	return v
}

// ValidateURL validates a URL to prevent SSRF attacks
func (v *URLValidator) ValidateURL(urlStr string) error {
	// If protection is disabled, allow all URLs
	if !v.config.Enabled {
		return nil
	}

	// Check for empty URL
	if urlStr == "" {
		return fmt.Errorf("URL is required")
	}

	// Parse URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	// Validate scheme
	if err := v.validateScheme(parsedURL); err != nil {
		return err
	}

	// Validate hostname
	hostname := parsedURL.Hostname()
	if hostname == "" {
		return fmt.Errorf("hostname is required")
	}

	// Check if hostname is an IP address directly
	if ip := net.ParseIP(hostname); ip != nil {
		return v.validateIP(ip)
	}

	// Resolve hostname to IPs
	if err := v.resolveAndValidate(hostname); err != nil {
		return err
	}

	return nil
}

// validateScheme checks if the URL scheme is allowed
func (v *URLValidator) validateScheme(parsedURL *url.URL) error {
	scheme := strings.ToLower(parsedURL.Scheme)
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("only http and https schemes allowed")
	}
	return nil
}

// resolveAndValidate resolves a hostname and validates all resolved IPs
func (v *URLValidator) resolveAndValidate(hostname string) error {
	// Normalize hostname to lowercase for comparison
	hostname = strings.ToLower(hostname)

	// Block localhost variations
	if hostname == "localhost" || strings.HasSuffix(hostname, ".localhost") {
		return fmt.Errorf("blocked hostname: %s", hostname)
	}

	// Resolve hostname to IPs
	ips, err := net.LookupIP(hostname)
	if err != nil {
		return fmt.Errorf("failed to resolve hostname: %w", err)
	}

	if len(ips) == 0 {
		return fmt.Errorf("no IP addresses found for hostname: %s", hostname)
	}

	// Validate each resolved IP
	for _, ip := range ips {
		if err := v.validateIP(ip); err != nil {
			return fmt.Errorf("blocked IP address %s for hostname %s: %w", ip, hostname, err)
		}
	}

	return nil
}

// validateIP validates that an IP address is not blocked
func (v *URLValidator) validateIP(ip net.IP) error {
	// Check if IP is in allowed networks (takes precedence)
	for _, network := range v.allowedNetworks {
		if network.Contains(ip) {
			return nil // Explicitly allowed
		}
	}

	// Check if IP is blocked
	if v.isBlockedIP(ip.String()) {
		return fmt.Errorf("blocked IP address: %s", ip)
	}

	// Check custom blocked networks
	for _, network := range v.blockedNetworks {
		if network.Contains(ip) {
			return fmt.Errorf("blocked IP address: %s", ip)
		}
	}

	return nil
}

// isBlockedIP checks if an IP address is in a blocked range
func (v *URLValidator) isBlockedIP(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		// If we can't parse it, block it to be safe
		return true
	}

	// Default blocked CIDR ranges
	blockedRanges := []string{
		"127.0.0.0/8",    // Loopback IPv4
		"10.0.0.0/8",     // Private network (RFC 1918)
		"172.16.0.0/12",  // Private network (RFC 1918)
		"192.168.0.0/16", // Private network (RFC 1918)
		"169.254.0.0/16", // Link-local (AWS/GCP metadata service!)
		"::1/128",        // IPv6 loopback
		"fc00::/7",       // IPv6 unique local addresses
		"fe80::/10",      // IPv6 link-local
	}

	for _, cidr := range blockedRanges {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		if network.Contains(ip) {
			return true
		}
	}

	return false
}

// ValidateURLWithLogging validates a URL and logs blocked attempts
// This is useful for security monitoring and audit trails
func (v *URLValidator) ValidateURLWithLogging(urlStr string, logger func(msg string, fields map[string]interface{})) error {
	err := v.ValidateURL(urlStr)
	if err != nil && logger != nil {
		logger("SSRF protection blocked URL", map[string]interface{}{
			"url":   urlStr,
			"error": err.Error(),
		})
	}
	return err
}
