# GoCamNet

A Go executable that demonstrates CLI argument parsing and timing functionality.

## Features

- Counts from 1 to N seconds (where N is a CLI parameter)
- Displays count every second
- Command-line argument parsing with both long and short forms
- Input validation

## Usage

This project uses a `Makefile` to simplify common operations.

```bash
# Build the executable
make build

# Run with default 10 seconds
make run

# Run with custom number of seconds
make run seconds=5
make run s=15

# Run the OpenCV example
make run-opencv

# Show help for CLI flags
./gocamnet -h
```

## Setup (macOS)

To set up the development environment and install dependencies like OpenCV 4, run the macOS setup target:

```bash
make install
```

## Development

```bash
# Build the executable for current platform
make build

# Run the executable
make run

# Run the OpenCV example
make run-opencv

# Build for different platforms (example)
GOOS=linux GOARCH=amd64 make build
GOOS=windows GOARCH=amd64 make build
```

## Testing

```bash
# Run all tests with verbose output
make test

# Run tests with coverage
make test -cover
```

## Project Structure

```
gocamnet/
├── main.go          # Main application entry point
├── main_test.go     # Unit tests for main package
├── go.mod           # Go module definition
├── .gitignore       # Git ignore patterns
├── README.md        # This file
├── Makefile         # Project build, run, and test commands
└── setup-mac.sh     # Script to install macOS dependencies
```