BINARY := md-docx
VERSION := $(shell tr -d '[:space:]' < versions.txt)

.PHONY: build clean bump-patch bump-minor bump-major

build:
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -X main.version=$(VERSION)" -o $(BINARY) ./cmd/md-docx

clean:
	rm -f $(BINARY)

bump-patch:
	@v=$$(tr -d '[:space:]' < versions.txt); MAJ=$${v%%.*}; rest=$${v#*.}; MIN=$${rest%%.*}; PAT=$${rest##*.}; printf '%s.%s.%s\n' "$$MAJ" "$$MIN" "$$((PAT + 1))" > versions.txt; cat versions.txt

bump-minor:
	@v=$$(tr -d '[:space:]' < versions.txt); MAJ=$${v%%.*}; rest=$${v#*.}; MIN=$${rest%%.*}; printf '%s.%s.0\n' "$$MAJ" "$$((MIN + 1))" > versions.txt; cat versions.txt

bump-major:
	@v=$$(tr -d '[:space:]' < versions.txt); MAJ=$${v%%.*}; printf '%s.0.0\n' "$$((MAJ + 1))" > versions.txt; cat versions.txt
