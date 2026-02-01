package parspackip

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/netip"
	"strings"
	"sync"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"go.uber.org/zap"
)

const (
	ipv4URL = "https://parspack.com/cdnips.txt"
)

func init() {
	caddy.RegisterModule(ParspackIPRange{})
}

// ParspackIPRange retrieves ParsPack CDN IP ranges from their official sources
type ParspackIPRange struct {
	// Interval specifies how often to refresh the IP list
	Interval caddy.Duration `json:"interval,omitempty"`

	// Timeout specifies the maximum time to wait for a response
	Timeout caddy.Duration `json:"timeout,omitempty"`

	logger   *zap.Logger
	ipRanges []netip.Prefix
	mu       sync.RWMutex
	stop     chan struct{}
}

// CaddyModule returns the Caddy module information
func (ParspackIPRange) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.ip_sources.parspack",
		New: func() caddy.Module { return new(ParspackIPRange) },
	}
}

// Provision implements caddy.Provisioner
func (p *ParspackIPRange) Provision(ctx caddy.Context) error {
	p.logger = ctx.Logger(p)

	// Set default interval if not specified
	if p.Interval == 0 {
		p.Interval = caddy.Duration(1 * time.Hour)
	}

	// Start background refresh
	p.stop = make(chan struct{})
	go p.refreshLoop()

	return nil
}

// GetIPRanges implements caddyhttp.IPRangeSource
func (p *ParspackIPRange) GetIPRanges(_ *http.Request) []netip.Prefix {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.ipRanges
}

// fetchIPRanges fetches IP ranges from ParsPack endpoint
func (p *ParspackIPRange) fetchIPRanges() error {
	ranges, err := p.fetchFromURL(ipv4URL)
	if err != nil {
		return fmt.Errorf("failed to fetch IPv4 ranges: %w", err)
	}

	p.mu.Lock()
	p.ipRanges = ranges
	p.mu.Unlock()

	p.logger.Info("successfully fetched IP ranges", zap.Int("count", len(ranges)))
	return nil
}

// fetchFromURL fetches IP ranges from a URL
func (p *ParspackIPRange) fetchFromURL(url string) ([]netip.Prefix, error) {
	ctx := context.Background()
	if p.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(p.Timeout))
		defer cancel()
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return p.parseIPRanges(string(body))
}

// parseIPRanges parses IP ranges from text (one per line, CIDR format)
func (p *ParspackIPRange) parseIPRanges(text string) ([]netip.Prefix, error) {
	var ranges []netip.Prefix
	lines := strings.Split(text, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		prefix, err := caddyhttp.CIDRExpressionToPrefix(line)
		if err != nil {
			p.logger.Warn("failed to parse IP range", zap.String("range", line), zap.Error(err))
			continue
		}

		ranges = append(ranges, prefix)
	}

	return ranges, nil
}

// refreshLoop periodically refreshes the IP ranges
func (p *ParspackIPRange) refreshLoop() {
	// First time fetch
	if err := p.fetchIPRanges(); err != nil {
		p.logger.Warn("failed to fetch initial IP ranges", zap.Error(err))
	}

	ticker := time.NewTicker(time.Duration(p.Interval))
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := p.fetchIPRanges(); err != nil {
				p.logger.Error("failed to refresh IP ranges", zap.Error(err))
			}
		case <-p.stop:
			return
		}
	}
}

// Cleanup implements caddy.CleanerUpper
func (p *ParspackIPRange) Cleanup() error {
	if p.stop != nil {
		close(p.stop)
	}
	return nil
}

// UnmarshalCaddyfile implements caddyfile.Unmarshaler
func (p *ParspackIPRange) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	d.Next() // Skip module name

	// No same-line options are supported
	if d.NextArg() {
		return d.ArgErr()
	}

	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "interval":
			if !d.NextArg() {
				return d.ArgErr()
			}
			dur, err := caddy.ParseDuration(d.Val())
			if err != nil {
				return d.Errf("invalid interval duration: %v", err)
			}
			p.Interval = caddy.Duration(dur)

		case "timeout":
			if !d.NextArg() {
				return d.ArgErr()
			}
			dur, err := caddy.ParseDuration(d.Val())
			if err != nil {
				return d.Errf("invalid timeout duration: %v", err)
			}
			p.Timeout = caddy.Duration(dur)

		default:
			return d.ArgErr()
		}
	}

	return nil
}

// Interface guards
var (
	_ caddy.Provisioner       = (*ParspackIPRange)(nil)
	_ caddy.CleanerUpper      = (*ParspackIPRange)(nil)
	_ caddyfile.Unmarshaler   = (*ParspackIPRange)(nil)
	_ caddyhttp.IPRangeSource = (*ParspackIPRange)(nil)
)
