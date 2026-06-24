# Gates. Doc-Gate via d-check.mk (erzeugt mit `d-check --print-mk`, v0.29.0 —
# bringt doc-check/doc-trace/doc-complete/doc-doctor/doc-repair/doc-help,
# netzlos via --network none). Code-Gates (lint/test/build) sind
# repo-spezifisch und wachsen mit dem Code (ADR-0002, Welle 1). Nur
# existierende, laufende Targets in AGENTS.md / harness/README.md eintragen
# (keine halluzinierten Gates, Modul 13).
include d-check.mk

.PHONY: help gates
help: ## Targets anzeigen
	@grep -hE '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-16s %s\n", $$1, $$2}'

gates: doc-check ## alle aktuell lauffähigen Gates (Code-Gates ergänzen)
