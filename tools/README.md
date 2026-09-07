# Tailwind CSS UI tooling
#
# This project uses the Tailwind **standalone CLI** (no Node/npm).
#
# ## Install
#
#     make tools
#
# Downloads a pinned binary into `.tools/tailwindcss` (gitignored) and verifies
# the SHA-256 against the official release `sha256sums.txt`.
#
# ## Build CSS
#
#     make css          # minify → web/static/css/app.css (committed)
#     make css-watch    # rebuild on change
#     make css-check    # CI: fail if committed CSS is stale
#
# Source: `web/styles/input.css`
# Output: `web/static/css/app.css`
#
# ## htmx
#
#     make htmx
#
# Vendors `web/static/js/htmx.min.js` at the version pinned in `versions.env`.
#
# ## Bumping versions
#
# Edit `versions.env`, then re-run `make tools` / `make htmx`. For htmx, set
# `HTMX_SHA256` to the printed actual hash after the first download.
