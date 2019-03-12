all: binary

.PHONY: binary
binary: ## build executable for Linux
	@echo "WARNING: binary creates a Linux executable. Use cross for macOS or Windows."
	./build/binary

.PHONY: upx
upx:
	./build/upx

.PHONY: clean
clean:
	rm -rf ./dist/*

.PHONY: binary-upx
binary-upx:
	./build/binary
	./build/upx
