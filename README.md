# GoCamNet

Simple Camera -> Neural Network -> Display program written in Golang, with OpenCV.  Calls a separate python process for the NN.

To try it out:

```
make install
make run-camera
```

## Usage

This project uses a `Makefile` to simplify common operations. The Go program manages a local PyTorch detector (YOLOv8n) via a lightweight TCP IPC; no TensorFlow/ONNX/Caffe is used.

```bash
# Install Python env (uv) and sync dependencies, then build
make install && make build

# Generate a sample image
make run-sample-image

# Display the primary camera feed (press any key to exit). The Go binary will spawn the PyTorch process automatically.
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

This project uses GoCV, which requires OpenCV to be installed on your system. Please see `setup-mac.sh` for macOS instructions. PyTorch dependencies are managed by `uv` from the `py/pyproject.toml`.

Build the project using the Makefile:
```sh
make build
```

## Testing

Unit tests for basic OpenCV operations and image generation remain. DNN tests were removed with the migration to a managed PyTorch subprocess.

## Project Structure

```
gocamnet/
├── cmd/gocamnet/main.go   # Main application entry point
├── main_test.go           # Unit tests for sample image & camera utilities
├── internal/pydetect/     # Go-managed PyTorch subprocess and IPC client
├── py/                    # Python inference server (YOLOv8n) managed by uv
├── go.mod           # Go module definition
├── .gitignore       # Git ignore patterns
├── README.md        # This file
├── AGENT.md         # Instructions for AI agents
├── Makefile         # Project build, run, and test commands
└── setup-mac.sh     # Script to install macOS dependencies
```

## For AI Agents

**IMPORTANT**: All AI agents working on this project must read the [AGENT.md](AGENT.md) file before making any changes. Always build via the Makefile so CGO/OpenCV env is set, and use `make install` to provision the Python `uv` environment under `py/`.


# Running on a remote Rasperry Pi

If you want to run this as a dashcam, you might add the following to your `.git/config` file:

```
[remote "dashcampi"]
    url = dashcampi:~/dev/gocamnet/
    fetch = +refs/heads/*:refs/remotes/dashcampi/*
```

Then on the pi:

```
mkdir -p ~/dev/gocamnet/
cd ~/dev/gocamnet/
git init
# Allow remote pushes to update the checked-out branch
git config receive.denyCurrentBranch updateInstead
```

Then you can push your changes to the remote repository:

```
git push dashcampi main
```