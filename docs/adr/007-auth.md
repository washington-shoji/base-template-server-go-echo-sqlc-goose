# ADR 007: Authentication

## Decision

Browser: HTTP-only session cookies + CSRF on `/app`. API: opaque bearer tokens. Common `authn.Principal` in context. Authorisation enforced in services.
