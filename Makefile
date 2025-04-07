LINTER = golangci-lint
LINTER_FLAGS = run

.DEFAULT_GOAL := lint

.PHONY: lint
lint:
	$(LINTER) $(LINTER_FLAGS)

.PHONY: lint-fix
lint-fix:
	$(LINTER) $(LINTER_FLAGS) --fix

.PHONY: install-linter
install-linter:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest

.PHONY: run-it-clean
run-it:
	go clean -testcache
	go test ./...

.PHONY: run-compose
run-compose:
	docker-compose up -d

.PHONY: run-compose-b
run-compose-b:
	docker-compose up -d --build