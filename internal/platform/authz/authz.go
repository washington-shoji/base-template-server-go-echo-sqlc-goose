package authz

import (
	"errors"

	"go-echo-server-template/internal/platform/authn"
)

var ErrPermissionDenied = errors.New("permission denied")

// RequireAuthenticated ensures a non-anonymous principal when auth is enabled.
func RequireAuthenticated(p authn.Principal, authDisabled bool) error {
	if authDisabled {
		return nil
	}
	if p.UserID == "" {
		return ErrPermissionDenied
	}
	return nil
}

func HasRole(p authn.Principal, role string) bool {
	for _, r := range p.Roles {
		if r == role {
			return true
		}
	}
	return false
}
