all: binary

.PHONY: binary
binary: ## build executable for Linux
	@echo "WARNING: binary creates a Linux executable. Use cross for macOS or Windows."
	./scripts/binary

.PHONY: upx
upx:
	./scripts/upx

.PHONY: clean
clean:
	rm -rf ./dist/*

.PHONY: binary-upx
binary-upx:
	./scripts/binary
	./scripts/upx
