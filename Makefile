# Gates. Doc-Gate via d-check.mk (erzeugt mit `d-check --print-mk`, v0.91.0 —
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

.PHONY: help gates lint test coverage-gate build verify-clean deps deps-verify self-check image image-test feature-test dist
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

# Mindestabdeckung in Prozent (ADR-0005): paketübergreifend über die Produktionspakete.
# Anheben ist ein gewöhnlicher PR; unter 90 lässt das Skript nicht zu (neue ADR nötig).
# Kein ?=: die Umgebung soll den Wert nicht überschreiben; nur ein ausdrücklicher Aufruf (make COVERAGE_MIN=...).
COVERAGE_MIN = 90

coverage-gate: ## Testabdeckung gegen COVERAGE_MIN prüfen (offline, ADR-0005)
	@mkdir -p .tmp/gocache
	$(GO_RUN) sh tools/coverage-gate.sh $(COVERAGE_MIN)

# Build-Informationen und reproduzierbare Flags (LH-QA-004): statisch gelinkt,
# ohne Pfade/Zeitstempel/VCS-Stempel im Binary.
VERSION ?= dev
COMMIT  ?= $(shell git rev-parse --short=12 HEAD 2>/dev/null || echo unknown)
CLI_PKG  = github.com/pt9912/bedrock-eu-guard/internal/adapter/driving/cli
LDFLAGS  = -s -w -X $(CLI_PKG).version=$(VERSION) -X $(CLI_PKG).commit=$(COMMIT)

build: ## Binary nach bin/bedrock-eu-check bauen (VERSION=..., COMMIT=...)
	@mkdir -p .tmp/gocache bin
	$(GO_RUN) env CGO_ENABLED=0 go build -trimpath -buildvcs=false -ldflags "$(LDFLAGS)" -o bin/bedrock-eu-check ./cmd/bedrock-eu-check

# Kontrollierte Umgebung für den Self-Check (EU-konforme Minimalumgebung, kein HOME).
SELF_ENV = env -i AWS_REGION=eu-central-1 CLAUDE_CODE_USE_BEDROCK=1 HOME=

self-check: build ## bedrock-eu-check gegen das eigene Repository (local, terraform); Exit 0 Pflicht (LH-QA-005)
	$(SELF_ENV) ./bin/bedrock-eu-check local --root . --ci
	$(SELF_ENV) ./bin/bedrock-eu-check terraform --root . --ci

IMAGE ?= bedrock-eu-check

image: ## Docker-Images bauen (slim, full); braucht Netz zum Ziehen der Basis-Images
	docker build --target slim --build-arg VERSION=$(VERSION) --build-arg COMMIT=$(COMMIT) -t $(IMAGE):slim .
	docker build --target full --build-arg VERSION=$(VERSION) --build-arg COMMIT=$(COMMIT) -t $(IMAGE):full .

image-test: image ## Smoke-Tests der Images (Laufzeit ohne Netz); nicht Teil von make gates
	tools/image-smoke.sh $(IMAGE):slim slim
	tools/image-smoke.sh $(IMAGE):full full

DIST_ARCHS = amd64 arm64

dist: ## Release-Artefakte (linux amd64/arm64) und SHA256SUMS nach dist/v$(VERSION)/ bauen; veroeffentlicht nichts (VERSION=x.y.z Pflicht)
	@case "$(VERSION)" in dev|"") echo "VERSION=x.y.z angeben (Release-Version)"; exit 1;; esac
	@mkdir -p .tmp/gocache dist/v$(VERSION)
	@rm -f dist/v$(VERSION)/bedrock-eu-check_* dist/v$(VERSION)/SHA256SUMS
	@for arch in $(DIST_ARCHS); do \
		echo "bauen: linux/$$arch"; \
		$(GO_RUN) env CGO_ENABLED=0 GOOS=linux GOARCH=$$arch go build -trimpath -buildvcs=false \
			-ldflags "$(LDFLAGS)" -o dist/v$(VERSION)/bedrock-eu-check_$(VERSION)_linux_$$arch ./cmd/bedrock-eu-check || exit 1; \
	done
	@cd dist/v$(VERSION) && sha256sum bedrock-eu-check_$(VERSION)_linux_* > SHA256SUMS && sha256sum -c SHA256SUMS

feature-test: ## Devcontainer-Feature gegen ein lokal gebautes Release testen (Docker, ohne Netz zur Laufzeit)
	GO_IMAGE=$(GO_IMAGE) tools/feature-test.sh

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

gates: doc-check doc-targets lint test coverage-gate ## alle lauffähigen Gates (doc-check, doc-targets, lint, test, coverage-gate)
