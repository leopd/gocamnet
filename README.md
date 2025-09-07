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

This project uses GoCV, which requires OpenCV to be installed on your system. Please see `setup-mac.sh` for macOS instructions.

Build the project using the Makefile:
```sh
make build
```

## Testing

The test suite includes a Deep Neural Network (DNN) based test that requires a pre-trained model. Due to the instability of public model download links, you must download the required model files manually.

**To run the full test suite:**

1.  **Download Model Files:**
    *   The required model architecture file, `MobileNetSSD_deploy.prototxt`, is included in this repository in the `models/` directory.
    *   You only need to download the corresponding model weights file: `MobileNetSSD_deploy.caffemodel`.
    *   A reliable source for this file is the repository at: `https://github.com/chuanqi305/MobileNet-SSD`

2.  **Place Files:**
    *   Ensure the `models/` directory contains both `MobileNetSSD_deploy.prototxt` (included) and `MobileNetSSD_deploy.caffemodel` (which you downloaded).

3.  **Run Tests:**
    Once both model files are in place, run the tests using the Makefile:
    ```sh
    make test
    ```

Without the model files, the DNN test will be automatically skipped.

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