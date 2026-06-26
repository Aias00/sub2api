package service

import "testing"

func TestShouldUseImplicitSMTPTLS(t *testing.T) {
	tests := []struct {
		name   string
		config *SMTPConfig
		want   bool
	}{
		{name: "implicit TLS submission port", config: &SMTPConfig{Port: 465, UseTLS: true}, want: true},
		{name: "STARTTLS submission port", config: &SMTPConfig{Port: 587, UseTLS: true}, want: false},
		{name: "alternate STARTTLS submission port", config: &SMTPConfig{Port: 2525, UseTLS: true}, want: false},
		{name: "TLS disabled", config: &SMTPConfig{Port: 465, UseTLS: false}, want: false},
		{name: "nil config", config: nil, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldUseImplicitSMTPTLS(tt.config); got != tt.want {
				t.Fatalf("shouldUseImplicitSMTPTLS() = %v, want %v", got, tt.want)
			}
		})
	}
}
