BIN_DIR := bin
CMD_DIR := cmd/git-ai-commit
BIN_NAME := git-ai-commit

.PHONY: build
build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(BIN_NAME) ./$(CMD_DIR)

.PHONY: test
test:
	go test ./...
