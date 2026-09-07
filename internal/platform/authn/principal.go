package authn

import (
	"context"
)

type contextKey string

const principalKey contextKey = "principal"

// Principal is the authenticated identity shared by web and API adapters.
type Principal struct {
	UserID     string
	Roles      []string
	AuthMethod string // cookie | bearer | disabled
}

func WithPrincipal(ctx context.Context, p Principal) context.Context {
	return context.WithValue(ctx, principalKey, p)
}

func FromContext(ctx context.Context) (Principal, bool) {
	p, ok := ctx.Value(principalKey).(Principal)
	return p, ok
}

func Anonymous() Principal {
	return Principal{UserID: "", Roles: nil, AuthMethod: "anonymous"}
}
