# Makefile for pf-tui-go

.PHONY: all help native mac win clean

# Define the target binary names
TARGET_NATIVE := pf-tui
TARGET_MAC_ARM64 := pf-tui.mac-arm64
TARGET_WIN := pf-tui.exe

help:
	@echo "Makefile for pf-tui-go"
	@echo ""
	@echo "Usage:"
	@echo "  make           Show this help message"
	@echo "  make all       Build for all supported platforms (Windows and macOS)"
	@echo "  make native    Build for the native OS"
	@echo "  make win       Build for Windows (64-bit)"
	@echo "  make mac       Build for macOS (Apple Silicon)"
	@echo "  make clean     Remove build artifacts"

# Build for all supported platforms
all: mac
#all: win mac

# Build for the native OS
build:
	go build -o $(TARGET_NATIVE)

# Build for macOS (Apple Silicon)
mac:
	GOOS=darwin GOARCH=arm64 go build -o $(TARGET_MAC_ARM64)

## Build for Windows (64-bit)
#win:
#	GOOS=windows GOARCH=amd64 go build -o $(TARGET_WIN)
#
# Clean up build artifacts
clean:
	rm -f $(TARGET_NATIVE) $(TARGET_MAC_ARM64) $(TARGET_WIN) 

