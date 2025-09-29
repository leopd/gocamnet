# GoCamNet

Simple Camera -> Neural Network -> Display program written in Golang, with OpenCV.  Calls a separate python process for the NN.

To try it out:

```
make install
make build
make run-camera
```

# Setting up Dash-cam on a Raspberry Pi

1. Install the dependencies:

```
make install
```

2. Create a desktop file:

```
mkdir -p ~/.config/autostart
ln -s $HOME/dev/gocamnet/dashcam.desktop ~/.config/autostart/dashcam.desktop
```

3. Start the dashcam:

```
./gocamnet --show-camera
```


## Usage

This project uses a `Makefile` to simplify common operations. The Go program manages a local PyTorch detector (YOLOv8n) via a lightweight TCP IPC; no TensorFlow/ONNX/Caffe is used.

```bash
# List available cameras
make list-cameras

# Use a specific camera (e.g., camera 1)
./gocamnet --show-camera --camera 1

# Show help for CLI flags
./gocamnet -h
```

## Install (macOS and Raspberry Pi)

Use a single command to provision OS dependencies (OpenCV, build tools), install Go and `uv`, and sync Python deps:

```bash
make install
```

Advanced usage (optional granular targets):

```bash
make install-os    # Detects OS and runs setup-mac.sh or setup-rpi.sh
make install-go    # Ensures Go >= go.mod version is installed
make install-uv    # Ensures uv is installed
make build         # Builds the Go binary
```

If you are on a headless system, `--show-camera` requires a desktop/X11 session. You can still validate the pipeline with:

```bash
PYDETECT_MOCK=1 ./gocamnet --show-camera
```

### Using the Pi's primary display over SSH

When you SSH into a Raspberry Pi that is connected to a monitor, Qt will fail to open a window unless we point it at the console session. To reuse the primary display:

```bash
export DISPLAY=:0
export XAUTHORITY=$HOME/.Xauthority
./gocamnet --show-camera
```

## Development

This project uses GoCV, which requires OpenCV to be installed on your system. `make install` will provision OpenCV and prerequisites via OS-specific scripts. PyTorch dependencies are managed by `uv` from the `py/pyproject.toml`.

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
├── setup-mac.sh     # Script to install macOS dependencies
└── setup-rpi.sh     # Script to install Raspberry Pi dependencies (Ubuntu/Debian)
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

# Troubleshooting

See if the go bindings to opencv are working:

```
make run-sample-image
```

If that works, see if the camera is working:

```
make run-camera
```