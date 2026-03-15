GOROOT ?= $(shell go env GOROOT)

# wasm_exec.js moved from misc/wasm/ to lib/wasm/ in Go 1.21
WASM_EXEC_CANDIDATES := \
	$(GOROOT)/lib/wasm/wasm_exec.js \
	$(GOROOT)/misc/wasm/wasm_exec.js

.PHONY: build dev clean

# Build WASM and copy wasm_exec.js — used by CI before npm publish
build:
	GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o csvlinter.wasm ../cmd/wasm
	@for f in $(WASM_EXEC_CANDIDATES); do \
		if [ -f "$$f" ]; then cp "$$f" wasm_exec.js && echo "Copied $$f"; break; fi; \
	done
	@[ -f wasm_exec.js ] || (echo "ERROR: wasm_exec.js not found in GOROOT=$(GOROOT)" && exit 1)

# Build and copy artifacts into the sibling csvlinter-app for local development
dev: build
	@mkdir -p ../../csvlinter-app/public
	cp csvlinter.wasm ../../csvlinter-app/public/csvlinter.wasm
	cp wasm_exec.js ../../csvlinter-app/public/wasm_exec.js
	@echo "Copied WASM artifacts to csvlinter-app/public/"

clean:
	rm -f csvlinter.wasm wasm_exec.js
	rm -rf dist/
