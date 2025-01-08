GOPATH ?= $(shell go env GOPATH)/bin
PATH := $(GOPATH):$(PATH)
PACKAGE := $(shell grep -Eo "module (.+)" go.mod | sed 's/module //g')

include google-oauth-vars.mk

ifeq ($(GOOGLE_CLIENT_ID),)
	$(error GOOGLE_CLIENT_ID is not set)
endif
ifeq ($(GOOGLE_CLIENT_SECRET),)
	$(error GOOGLE_CLIENT_SECRET is not set)
endif

.PHONY: gobins
gobins:
	go install golang.org/x/tools/cmd/stringer@latest
	go install github.com/air-verse/air@latest
	go install github.com/a-h/templ/cmd/templ@v0.2.793

.PHONY: deps
deps:
	go get ./...

.PHONY: generate
generate:
	go generate ./...

# run templ generation in watch mode to detect all .templ files and 
# re-create _templ.txt files on change, then send reload event to browser. 
# Default url: http://localhost:7331
live/templ:
	${GOPATH}/templ generate --watch --proxy="http://localhost:8080" --open-browser=false -v

# run air to detect any go file changes to re-build and re-run the server.
live/server:
	${GOPATH}/air \
	--build.cmd "go build -ldflags \"-X ${PACKAGE}/pkg/auth.GOOGLE_CLIENT_ID=${GOOGLE_CLIENT_ID} -X ${PACKAGE}/pkg/auth.GOOGLE_CLIENT_SECRET=${GOOGLE_CLIENT_SECRET}\" -o tmp/bin/main" --build.bin "tmp/bin/main" --build.delay "100" \
	--build.exclude_dir "node_modules" \
	--build.include_ext "go" \
	--build.stop_on_error "false" \
	--misc.clean_on_exit true

live/htmx:
	cp ./node_modules/htmx.org/dist/htmx.min.js ./assets/htmx.min.js

# run tailwindcss to generate the styles.css bundle in watch mode.
live/tailwind:
	npx --yes tailwindcss -i ./input.css -o ./assets/styles.css --minify --watch

# watch for any js or css change in the assets/ folder, then reload the browser via templ proxy.
live/sync_assets:
	${GOPATH}/air \
	--build.cmd "${GOPATH}/templ generate --notify-proxy" \
	--build.bin "true" \
	--build.delay "100" \
	--build.exclude_dir "" \
	--build.include_dir "assets" \
	--build.include_ext "js,css"

# start all 5 watch processes in parallel.
.PHONY: live
live: 
	make generate
	make -j5 live/templ live/server live/tailwind live/htmx live/sync_assets

# Build the whole thing
.PHONY: build
build:
	make gobins
	go generate ./...
	make deps
	${GOPATH}/templ generate
	npx --yes tailwindcss -i ./input.css -o ./assets/styles.css --minify
	go build -ldflags "-s -w -X ${PACKAGE}/pkg/auth.GOOGLE_CLIENT_ID=${GOOGLE_CLIENT_ID} -X ${PACKAGE}/pkg/auth.GOOGLE_CLIENT_SECRET=${GOOGLE_CLIENT_SECRET}" -o build/bin/mailcode-v2

.PHONY: clean
clean:
	rm -rf tmp/
	rm -rf build/
	go clean -cache
	go clean -modcache