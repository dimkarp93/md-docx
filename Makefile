BINARY := md-docx
VERSION := $(shell tr -d '[:space:]' < versions.txt)
export GOWORK := off
export GOFLAGS := -mod=vendor

.PHONY: build test test-v test-run cover check clean bump-patch bump-minor bump-major

CHANNEL ?= local

build:
	@u=$$(git remote get-url origin 2>/dev/null || true); \
	case "$$u" in \
	  "")    o=local ;; \
	  *://*) h=$${u#*://}; h=$${h#*@}; o="https://$${h%.git}" ;; \
	  *:*)   h=$${u#*@};   o="https://$$(printf '%s' "$${h%.git}" | tr ':' '/')" ;; \
	  *)     o=local ;; \
	esac; \
	if [ -f upstream.txt ]; then up=$$(tr -d '[:space:]' < upstream.txt); else up="$$o"; fi; \
	c=$$(git rev-parse --short HEAD 2>/dev/null || true); \
	CGO_ENABLED=0 go build -trimpath \
	  -ldflags="-s -w -X main.version=$(VERSION) -X main.origin=$$o -X main.upstream=$$up -X main.commit=$$c -X main.channel=$(CHANNEL)" \
	  -o $(BINARY) ./cmd/md-docx

test:
	go test ./...

test-v:
	go test -v ./...

test-run:
	@test -n "$(T)" || { echo "specify a test: make test-run T=TestPageBreak"; exit 1; }
	go test -v -run '$(T)' ./...

cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | tail -1
	@echo "detailed report: go tool cover -html=coverage.out"

check: vet test

vet:
	go vet ./...

clean:
	rm -f $(BINARY)

bump-patch:
	@v=$$(tr -d '[:space:]' < versions.txt); MAJ=$${v%%.*}; rest=$${v#*.}; MIN=$${rest%%.*}; PAT=$${rest##*.}; printf '%s.%s.%s\n' "$$MAJ" "$$MIN" "$$((PAT + 1))" > versions.txt; cat versions.txt

bump-minor:
	@v=$$(tr -d '[:space:]' < versions.txt); MAJ=$${v%%.*}; rest=$${v#*.}; MIN=$${rest%%.*}; printf '%s.%s.0\n' "$$MAJ" "$$((MIN + 1))" > versions.txt; cat versions.txt

bump-major:
	@v=$$(tr -d '[:space:]' < versions.txt); MAJ=$${v%%.*}; printf '%s.0.0\n' "$$((MAJ + 1))" > versions.txt; cat versions.txt

vendor:
	GOWORK=off go mod tidy
	GOWORK=off go mod vendor

vendor-check:
	GOWORK=off go mod vendor
	test -z "$$(git status --porcelain -- go.mod go.sum vendor/ | tee /dev/stderr)"

.PHONY: vendor vendor-check
