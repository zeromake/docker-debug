all: binary

.PHONY: binary
binary: ## build executable for Linux
	@echo "WARNING: binary creates a Linux executable. Use cross for macOS or Windows."
	./scripts/binary.sh

.PHONY: binary-arm64
binary-arm64: ## build executable for Linux ARM 64-bit
	./scripts/binary.sh linux-arm64

.PHONY: upx
upx:
	./scripts/upx.sh

.PHONY: clean
clean:
	rm -rf ./dist/*

.PHONY: binary-upx
binary-upx:
	./scripts/binary.sh
	./scripts/upx.sh
