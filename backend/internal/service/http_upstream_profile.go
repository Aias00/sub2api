package service

import (
	"context"

	"github.com/Aias00/cloudbase/internal/gateway"
)

type HTTPUpstreamProfile = gateway.HTTPUpstreamProfile

const (
	HTTPUpstreamProfileDefault = gateway.HTTPUpstreamProfileDefault
	HTTPUpstreamProfileOpenAI  = gateway.HTTPUpstreamProfileOpenAI
)

// WithHTTPUpstreamProfile injects an upstream transport profile into ctx.
func WithHTTPUpstreamProfile(ctx context.Context, profile HTTPUpstreamProfile) context.Context {
	return gateway.WithHTTPUpstreamProfile(ctx, profile)
}

// HTTPUpstreamProfileFromContext resolves the upstream transport profile from ctx.
func HTTPUpstreamProfileFromContext(ctx context.Context) HTTPUpstreamProfile {
	return gateway.HTTPUpstreamProfileFromContext(ctx)
}
