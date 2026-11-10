# Gates. Doc-Gate via d-check.mk (erzeugt mit `d-check --print-mk`, v0.29.0 —
# bringt doc-check/doc-trace/doc-complete/doc-doctor/doc-repair/doc-help,
# netzlos via --network none). Code-Gates (lint/test/build) sind
# repo-spezifisch und wachsen mit dem Code (ADR-0002, Welle 1). Nur
# existierende, laufende Targets in AGENTS.md / harness/README.md eintragen
# (keine halluzinierten Gates, Modul 13).
include d-check.mk

# Go-Toolchain containerisiert (kein lokales go noetig), netzlos, Image per
# Digest gepinnt (Reproduzierbarkeit). Abhaengigkeiten kommen aus vendor/
# (ADR-0004); die Gates brauchen daher kein Netz.
GO_IMAGE ?= golang@sha256:699337d620559a59b4a2bb298ad59611e535d2ee755a34cf2d2a98f37578dc80
GO_RUN = docker run --rm --network none -u $(shell id -u):$(shell id -g) \
	-e HOME=/tmp -e GOCACHE=/src/.tmp/gocache -e GOFLAGS=-mod=vendor \
	-v "$(CURDIR):/src" -w /src $(GO_IMAGE)

# Nur fuer deps/deps-verify: mit Netz und beschreibbarem Modul-Cache (.tmp/ ist
# gitignored). Aendert vendor/, go.mod und go.sum; vendor/ nie von Hand editieren.
GO_NET = docker run --rm -u $(shell id -u):$(shell id -g) \
	-e HOME=/tmp -e GOCACHE=/src/.tmp/gocache -e GOMODCACHE=/src/.tmp/gomod \
	-e GOFLAGS=-mod=mod -v "$(CURDIR):/src" -w /src $(GO_IMAGE)

.PHONY: help gates lint test build verify-clean deps deps-verify
help: ## Targets anzeigen
	@grep -hE '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-16s %s\n", $$1, $$2}'

lint: ## gofmt-Check (ohne vendor/) + go vet
	@mkdir -p .tmp/gocache
	@out=$$($(GO_RUN) sh -c 'gofmt -l $$(find . -name "*.go" -not -path "./vendor/*" -not -path "./.tmp/*")'); \
	if [ -n "$$out" ]; then echo "gofmt-Befunde:"; echo "$$out"; exit 1; fi
	$(GO_RUN) go vet ./...

test: ## Unit-Tests (go test)
	@mkdir -p .tmp/gocache
	$(GO_RUN) go test ./...

build: ## Binary nach bin/bedrock-eu-check bauen
	@mkdir -p .tmp/gocache bin
	$(GO_RUN) go build -o bin/bedrock-eu-check ./cmd/bedrock-eu-check

deps: ## go mod tidy + vendor/ neu erzeugen (mit Netz, nicht Teil der Gates)
	@mkdir -p .tmp/gocache .tmp/gomod
	$(GO_NET) sh -c 'go mod tidy && go mod vendor'

deps-verify: ## go mod verify + vendor/ entspricht Upstream (mit Netz, nur bei Aenderungen an vendor/go.mod/go.sum)
	@mkdir -p .tmp/gocache .tmp/gomod
	$(GO_NET) sh -c 'go mod verify && go mod vendor'
	@out=$$(git status --porcelain -- vendor go.mod go.sum); \
	if [ -n "$$out" ]; then echo "vendor/go.mod/go.sum weichen von Upstream ab:"; echo "$$out"; exit 1; fi

verify-clean: ## gates in frischem git worktree des committeten HEAD (kein Arbeitsbaum-Ballast)
	@mkdir -p .tmp && d=$$(mktemp -d "$(CURDIR)/.tmp/verify.XXXXXX") && chmod 755 "$$d" && \
	git worktree add -q --detach "$$d/wt" HEAD && \
	$(MAKE) -C "$$d/wt" gates; rc=$$?; \
	git worktree remove --force "$$d/wt"; rm -rf "$$d"; exit $$rc

gates: doc-check lint test ## alle lauffähigen Gates (doc-check, lint, test)
