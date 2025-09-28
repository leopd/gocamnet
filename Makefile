NAME := gocamnet
OS := $(shell uname -s)
ARCH := $(shell uname -m)

EXECUTABLE := gocamnet

 

# Configure environment for macOS arm64 (Apple Silicon)
ifeq ($(OS),Darwin)
  ifeq ($(ARCH),arm64)
    export GOOS := darwin
    export GOARCH := arm64
    export CGO_ENABLED := 1
    export CC := clang
    export CXX := clang++
    export CGO_CFLAGS := -O2 -g -arch arm64
    export CGO_CXXFLAGS := -O2 -g -arch arm64
    export CGO_LDFLAGS := -O2 -g -arch arm64
    export PKG_CONFIG_PATH := $(shell brew --prefix opencv)/lib/pkgconfig:$(PKG_CONFIG_PATH)
  endif
endif

# Build dynamic CLI args from make variables (e.g., `make run seconds=5` or `make run s=5`)
RUN_ARGS :=
ifneq ($(seconds),)
  RUN_ARGS += -seconds $(seconds)
endif
ifneq ($(s),)
  RUN_ARGS += -s $(s)
endif

.PHONY: all build run run-sample-image run-camera list-cameras test test-go test-py clean install install-os install-go install-uv python-sync clean-models setup-os

all: build

build:
ifeq ($(OS), Darwin)
	go build -ldflags="-s -w" -o $(NAME) ./cmd/gocamnet
else
	CGO_ENABLED=1 CGO_CXXFLAGS="$(CGO_CXXFLAGS) -include opencv2/aruco.hpp" go build -o $(NAME) ./cmd/gocamnet
endif

run:
	./$(NAME) $(RUN_ARGS)

run-sample-image:
	./$(NAME) --sample-image $(RUN_ARGS)

run-camera:
	./$(NAME) --show-camera $(RUN_ARGS)

list-cameras:
	./$(NAME) --list-cameras $(RUN_ARGS)

test: test-go test-py

test-go:
	go test -v ./...

test-py: python-sync
	@cd py && uv run python -m pytest

install: install-os install-go install-uv python-sync

install-uv:
	@if ! command -v uv >/dev/null 2>&1; then \
	  if command -v brew >/dev/null 2>&1; then \
	    echo "Installing uv via Homebrew..."; brew install uv; \
	  else \
	    echo "Installing uv via official script..."; \
	    curl -LsSf https://astral.sh/uv/install.sh | sh; \
	  fi; \
	else \
	  echo "uv already installed: $$(uv --version)"; \
	fi

python-sync: install-uv
	@cd py && uv sync

clean:
	go clean
	 rm -f $(NAME)
	 rm -f output.jpg

# Optional: remove legacy models directory
clean-models:
	rm -rf models

install-os:
	@echo "Detecting OS for install..."
	@if [ "$(OS)" = "Darwin" ]; then \
		echo "Detected macOS"; \
		./setup-mac.sh; \
	elif [ "$(OS)" = "Linux" ]; then \
		if [ -f /proc/device-tree/model ] && grep -qi 'raspberry pi' /proc/device-tree/model; then \
			echo "Detected Raspberry Pi"; \
			./setup-rpi.sh; \
		elif uname -m | grep -Eq '^(armv7l|aarch64|arm64)$$'; then \
			echo "Detected Linux on ARM (assuming Raspberry Pi)"; \
			./setup-rpi.sh; \
		else \
			echo "Error: install-os supports only macOS or Raspberry Pi"; \
			exit 1; \
		fi; \
	else \
		echo "Error: install-os supports only macOS or Raspberry Pi"; \
		exit 1; \
	fi

# Back-compat alias; will be removed later
setup-os: install-os

install-go:
	@echo "Ensuring Go toolchain is installed..."
	@if command -v go >/dev/null 2>&1; then \
	  echo "Go already installed: $$(go version)"; \
	else \
	  if [ "$(OS)" = "Darwin" ] && command -v brew >/dev/null 2>&1; then \
	    echo "Installing Go via Homebrew..."; \
	    brew install go; \
	  elif [ "$(OS)" = "Linux" ]; then \
	    echo "Installing Go via apt..."; \
	    sudo apt-get update -y; \
	    sudo apt-get install -y golang-go; \
	  else \
	    echo "Please install Go manually for OS $(OS)"; \
	    exit 1; \
	  fi; \
	fi
