NAME := gocamnet
OS := $(shell uname -s)
ARCH := $(shell uname -m)

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

.PHONY: all build run run-sample-image run-camera list-cameras test clean install

all: build

build:
ifeq ($(OS), Darwin)
	go build -ldflags="-s -w" -o $(NAME)
else
	go build -o $(NAME)
endif

run:
	./$(NAME) $(RUN_ARGS)

run-sample-image:
	./$(NAME) --sample-image $(RUN_ARGS)

run-camera:
	./$(NAME) --show-camera $(RUN_ARGS)

list-cameras:
	./$(NAME) --list-cameras $(RUN_ARGS)

test:
	go test -v ./...

install:
ifeq ($(OS), Darwin)
	@echo "Running macOS setup script..."
	@./setup-mac.sh
else
	@echo "Installation for $(OS) is not yet implemented."
	@echo "Please install dependencies manually."
endif

clean:
	go clean
	rm -f $(NAME)
	rm -f output.jpg
