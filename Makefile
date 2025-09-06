NAME := gocamnet
OS := $(shell uname -s)

.PHONY: all build run run-opencv test clean install

all: build

build:
	go build -o $(NAME)

run:
	./$(NAME)

run-opencv:
	./$(NAME) --opencv

test:
	go test -v ./...

install:
ifeq ($(OS), Darwin)
	@echo "Running macOS setup script..."
	@chmod +x setup-mac.sh
	@./setup-mac.sh
else
	@echo "Installation for $(OS) is not yet implemented."
	@echo "Please install dependencies manually."
endif

clean:
	go clean
	rm -f $(NAME)
	rm -f output.jpg
