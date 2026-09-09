BINARY := md-docx
VERSION := $(shell tr -d '[:space:]' < versions.txt)

.PHONY: build test test-v test-run cover check clean configure unconfigure bump-patch bump-minor bump-major

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

test: hint
	go test ./...

test-v: hint
	go test -v ./...

test-run: hint
	@test -n "$(T)" || { echo "укажите тест: make test-run T=TestPageBreak"; exit 1; }
	go test -v -run '$(T)' ./...

cover: hint
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | tail -1
	@echo "детальный отчёт: go tool cover -html=coverage.out"

check: vet test

vet: hint
	go vet ./...

hint:
	@test -f go.work -o -f go.sum || echo "подсказка: md-libs недоступна — выполните make configure (нужна ../md-libs)"

clean:
	rm -f $(BINARY)

configure:
	@test -d ../md-libs || { echo "../md-libs не найден: клонируйте github.com/dimkarp93/md-libs рядом с этим репозиторием"; exit 1; }
	@test -d ../install-libs || { echo "../install-libs не найден: клонируйте github.com/dimkarp93/install-libs рядом с этим репозиторием"; exit 1; }
	cp go.work.local go.work
	go build -o /dev/null ./...

unconfigure:
	rm -f go.work go.work.sum

bump-patch:
	@v=$$(tr -d '[:space:]' < versions.txt); MAJ=$${v%%.*}; rest=$${v#*.}; MIN=$${rest%%.*}; PAT=$${rest##*.}; printf '%s.%s.%s\n' "$$MAJ" "$$MIN" "$$((PAT + 1))" > versions.txt; cat versions.txt

bump-minor:
	@v=$$(tr -d '[:space:]' < versions.txt); MAJ=$${v%%.*}; rest=$${v#*.}; MIN=$${rest%%.*}; printf '%s.%s.0\n' "$$MAJ" "$$((MIN + 1))" > versions.txt; cat versions.txt

bump-major:
	@v=$$(tr -d '[:space:]' < versions.txt); MAJ=$${v%%.*}; printf '%s.0.0\n' "$$((MAJ + 1))" > versions.txt; cat versions.txt
