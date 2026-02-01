package parspackip

import (
	"testing"

	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
)

func TestUnmarshalCaddyfile(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(*ParspackIPRange) error
	}{
		{
			name:  "minimal config",
			input: `parspack`,
			check: func(p *ParspackIPRange) error {
				// Should parse without error
				return nil
			},
		},
		{
			name: "full config",
			input: `parspack {
				interval 2h
				timeout 30s
			}`,
			check: func(p *ParspackIPRange) error {
				// Should parse without error
				return nil
			},
		},
		{
			name:    "invalid directive",
			input:   `parspack { invalid_option }`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &ParspackIPRange{}
			d := caddyfile.NewTestDispenser(tt.input)
			err := p.UnmarshalCaddyfile(d)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalCaddyfile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.check != nil && err == nil {
				if err := tt.check(p); err != nil {
					t.Errorf("check failed: %v", err)
				}
			}
		})
	}
}
