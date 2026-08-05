.PHONY: run test coverage

COVERAGE_DIR = coverage
COVER_OUT = $(COVERAGE_DIR)/coverage.out
COVER_HTML = $(COVERAGE_DIR)/coverage.html

run: 
	go run ./cmd/main.go

test:
	go test ./...

coverage:
	mkdir -p $(COVERAGE_DIR)
	go test ./... -p 1 -coverprofile=$(COVER_OUT)
	go tool cover -html=$(COVER_OUT) -o $(COVER_HTML)
	cmd.exe /c start chrome $$(wslpath -w $(COVER_HTML)) || true