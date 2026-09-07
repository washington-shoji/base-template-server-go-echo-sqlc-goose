# Tailwind CSS UI foundation tooling (no Node).
# Requires: curl, sha256sum, Go toolchain.

ROOT := $(abspath .)
TOOLS := $(ROOT)/.tools
TW := $(TOOLS)/tailwindcss

.PHONY: tools css css-watch css-check htmx run worker test build dev tidy help

help:
	@echo "Targets:"
	@echo "  make tools      Download Tailwind standalone CLI into .tools/"
	@echo "  make htmx       Download pinned htmx into web/static/js/"
	@echo "  make css        Build minified web/static/css/app.css"
	@echo "  make css-watch  Rebuild CSS on template/style changes"
	@echo "  make css-check  Fail if committed app.css is stale"
	@echo "  make run        Run HTTP server"
	@echo "  make worker     Run job worker"
	@echo "  make test       Run unit tests"
	@echo "  make build      Build CSS then Go binaries"
	@echo "  make dev        css-watch + server (two processes)"

tools:
	@bash tools/install-tailwind.sh

htmx:
	@bash tools/install-htmx.sh

css: $(TW)
	@$(TW) -i web/styles/input.css -o web/static/css/app.css --minify

css-watch: $(TW)
	@$(TW) -i web/styles/input.css -o web/static/css/app.css --watch

css-check: css
	@git diff --exit-code -- web/static/css/app.css || (echo "web/static/css/app.css is stale; commit make css output" >&2; exit 1)

$(TW):
	@$(MAKE) tools

run:
	go run ./cmd/server

worker:
	go run ./cmd/worker

test:
	go test ./internal/...

build: css
	go build -o bin/server ./cmd/server
	go build -o bin/worker ./cmd/worker

dev:
	@echo "Starting css-watch and server (Ctrl+C stops both)..."
	@$(MAKE) css-watch & TW_PID=$$!; \
	go run ./cmd/server; \
	kill $$TW_PID 2>/dev/null || true

tidy:
	go mod tidy && go mod vendor
