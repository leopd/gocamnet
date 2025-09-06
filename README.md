# GoCamNet

A Go executable that demonstrates OpenCV integration for computer vision tasks.

## Features

- Generate sample images using OpenCV
- Display the primary camera feed with optimized performance (30+ FPS)
- Real-time FPS monitoring and reporting
- Camera selection and discovery
- Command-line argument parsing with both long and short forms
- Input validation

## Usage

This project uses a `Makefile` to simplify common operations.

```bash
# Build the executable
make build

# Generate a sample image
make run-sample-image

# Display the primary camera feed (press any key to exit)
make run-camera

# List available cameras
make list-cameras

# Use a specific camera (e.g., camera 1)
./gocamnet --show-camera --camera 1

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

# Generate a sample image
make run-sample-image

# Display the primary camera feed
make run-camera

# List available cameras
make list-cameras

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
├── AGENT.md         # Instructions for AI agents
├── Makefile         # Project build, run, and test commands
└── setup-mac.sh     # Script to install macOS dependencies
```

## For AI Agents

**IMPORTANT**: All AI agents working on this project must read the [AGENT.md](AGENT.md) file before making any changes. This file contains critical instructions that must be followed for every modification to ensure code quality and proper documentation.