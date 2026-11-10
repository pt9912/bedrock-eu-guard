# Gates. Doc-Gate via d-check.mk (erzeugt mit `d-check --print-mk`, v0.29.0 —
# bringt doc-check/doc-trace/doc-complete/doc-doctor/doc-repair/doc-help,
# netzlos via --network none). Code-Gates (lint/test/build) sind
# repo-spezifisch und wachsen mit dem Code (ADR-0002, Welle 1). Nur
# existierende, laufende Targets in AGENTS.md / harness/README.md eintragen
# (keine halluzinierten Gates, Modul 13).
include d-check.mk

# Go-Toolchain containerisiert (kein lokales go noetig), netzlos, Image per
# Digest gepinnt (Reproduzierbarkeit). Stdlib-only => kein Netz noetig.
GO_IMAGE ?= golang@sha256:699337d620559a59b4a2bb298ad59611e535d2ee755a34cf2d2a98f37578dc80
GO_RUN = docker run --rm --network none -u $(shell id -u):$(shell id -g) \
	-e HOME=/tmp -e GOCACHE=/src/.tmp/gocache -e GOFLAGS=-mod=mod \
	-v "$(CURDIR):/src" -w /src $(GO_IMAGE)

.PHONY: help gates lint test build
help: ## Targets anzeigen
	@grep -hE '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-16s %s\n", $$1, $$2}'

lint: ## gofmt-Check + go vet
	@mkdir -p .tmp/gocache
	@out=$$($(GO_RUN) gofmt -l .); \
	if [ -n "$$out" ]; then echo "gofmt-Befunde:"; echo "$$out"; exit 1; fi
	$(GO_RUN) go vet ./...

test: ## Unit-Tests (go test)
	@mkdir -p .tmp/gocache
	$(GO_RUN) go test ./...

build: ## Binary nach bin/bedrock-eu-check bauen
	@mkdir -p .tmp/gocache bin
	$(GO_RUN) go build -o bin/bedrock-eu-check ./cmd/bedrock-eu-check

gates: doc-check lint test ## alle lauffähigen Gates (doc-check, lint, test)
