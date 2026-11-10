# bedrock-eu-check — mehrstufiger, reproduzierbarer Build (LH-TECH-004, LH-QA-004).
#
# Ziele:
#   slim  (Default) Nur das statische Binary — Modi local, terraform. Kein Shell, kein Root.
#   full            Zusätzlich die AWS-CLI (gepinnt) — alle Modi inkl. aws und all.
#
# Alle Basis-Images sind per Digest gepinnt. Der Build selbst braucht kein Netz
# (Abhängigkeiten liegen unter vendor/, ADR-0004); nur das Ziehen der Basis-
# Images braucht Netz.

FROM golang@sha256:699337d620559a59b4a2bb298ad59611e535d2ee755a34cf2d2a98f37578dc80 AS build
ARG VERSION=dev
ARG COMMIT=unknown
ENV CGO_ENABLED=0 GOFLAGS=-mod=vendor GOCACHE=/tmp/gocache
WORKDIR /src
COPY go.mod go.sum ./
COPY vendor ./vendor
COPY cmd ./cmd
COPY internal ./internal
RUN go build -trimpath -buildvcs=false \
      -ldflags "-s -w \
        -X github.com/pt9912/bedrock-eu-guard/internal/adapter/driving/cli.version=${VERSION} \
        -X github.com/pt9912/bedrock-eu-guard/internal/adapter/driving/cli.commit=${COMMIT}" \
      -o /out/bedrock-eu-check ./cmd/bedrock-eu-check

# AWS-CLI v2.39.7 (Amazon Linux 2023), per Digest gepinnt.
FROM amazon/aws-cli:2.39.7@sha256:bba262340b0bc7cbf3a8a48612531263b56f4b280c4f036a07816af1b271bda0 AS full
COPY --from=build /out/bedrock-eu-check /usr/local/bin/bedrock-eu-check
RUN mkdir -p /home/nonroot /work && chown 65532:65532 /home/nonroot /work
ENV HOME=/home/nonroot
USER 65532:65532
WORKDIR /work
ENTRYPOINT ["/usr/local/bin/bedrock-eu-check"]

FROM scratch AS slim
COPY --from=build /out/bedrock-eu-check /usr/local/bin/bedrock-eu-check
USER 65532:65532
WORKDIR /work
ENTRYPOINT ["/usr/local/bin/bedrock-eu-check"]
