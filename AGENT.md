# AGENT INSTRUCTIONS

## CRITICAL: ALL AI AGENTS MUST READ THIS FILE

This file contains essential instructions for any AI agent working on this GoCamNet project. **READ THIS FILE CLOSELY AND HONOR IT FOR EVERY CHANGE.**

## Project Overview

GoCamNet is a Go executable that demonstrates OpenCV integration for computer vision tasks, specifically focused on camera operations and image generation.

## Current Features

- **Sample Image Generation**: Generate sample images using OpenCV (`--sample-image`)
- **Camera Operations**: Display camera feeds and list available cameras
  - `--show-camera`: Display primary camera feed (press any key to exit)
  - `--list-cameras`: List available cameras and exit
  - `--camera <index>`: Specify which camera to use

## MANDATORY WORKFLOW

### After Every Set of Changes:
1. **ALWAYS run `make build`** to ensure the project compiles
2. **ALWAYS run `make test`** to ensure tests pass
3. **ALWAYS keep the README.md file up-to-date** with documentation about the current project

### Code Standards
- Follow Go best practices and conventions
- Maintain clean, readable code
- Add appropriate error handling
- Include helpful comments for complex logic

### Testing Requirements
- Add tests for new functionality
- Ensure existing tests continue to pass
- Test both success and error cases

### Documentation Requirements
- Update README.md for any new features or changes
- Keep usage examples current
- Document any new command-line flags
- Update the features list when adding/removing functionality

## Project Structure

```
gocamnet/
├── main.go          # Main application entry point
├── main_test.go     # Unit tests for main package
├── go.mod           # Go module definition
├── Makefile         # Project build, run, and test commands
├── README.md        # Project documentation
├── AGENT.md         # This file - agent instructions
└── setup-mac.sh     # macOS setup script
```

## Available Make Targets

- `make build` - Build the executable
- `make run-sample-image` - Generate a sample image
- `make run-camera` - Display camera feed
- `make list-cameras` - List available cameras
- `make test` - Run tests
- `make clean` - Clean build artifacts

## Important Notes

- This is a local project (not cloud-based)
- Runs on Mac, Windows, or Raspberry Pi (Ubuntu)
- Never use `chmod` at runtime - executable bit is stored in git
- Camera functionality requires proper permissions on macOS

## CGO/OpenCV environment: Always use the Makefile

- Always run builds and tests via the Makefile so required CGO/OpenCV environment variables are set.
- Do NOT run `go build` or `go test` directly. Use:
  - `make build`
  - `make test`
- Rationale: the Makefile exports `GOOS`, `GOARCH`, `CGO_ENABLED`, `CC`, `CXX`, `CGO_*FLAGS`, and sets `PKG_CONFIG_PATH` (e.g., to Homebrew's OpenCV on macOS arm64). Running `go test` directly may fail to link OpenCV even though `make test` works.

## REMINDER

**EVERY AI AGENT MUST:**
1. Read this file before making any changes
2. Run `make build` after every change
3. Run `make test` after every change
4. Keep README.md up-to-date
5. Follow the established patterns and conventions

**FAILURE TO FOLLOW THESE INSTRUCTIONS WILL RESULT IN BROKEN CODE AND POOR DOCUMENTATION.**

## Agent Instructions for GoCamNet Project

This document provides critical guidelines for AI agents working on the GoCamNet repository. Adherence to these rules is mandatory.

### CGO/OpenCV Environment: Always use the Makefile

**CRITICAL**: When working on macOS arm64 (Apple Silicon), the `Makefile` sets essential CGO environment variables (e.g., `CGO_CFLAGS`, `CGO_LDFLAGS`, `PKG_CONFIG_PATH`) that are required for GoCV to correctly link against your Homebrew-installed OpenCV libraries.

**DO NOT run `go build` or `go test` directly.** Always use the `make` targets:
- `make build`
- `make test`
- `make run`

Running `go build` or `go test` directly from the command line without these environment variables will result in linker errors because the Go compiler will not be able to find the necessary OpenCV libraries.

### Testing: NEVER Remove or Disable Failing Tests

**CRITICAL**: Under no circumstances should a failing test be removed, deleted, disabled, or skipped (e.g., using `t.Skipf`) as a method for making the test suite pass. A failing test indicates a regression or a flaw in the implementation. The only valid solution is to fix the underlying code or the test itself until it passes correctly. Removing a test to achieve a "green" build is a critical failure of duty and is strictly forbidden.
