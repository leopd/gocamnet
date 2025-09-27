#!/bin/bash

# This script installs system dependencies for GoCamNet on Raspberry Pi
# running Ubuntu/Debian (64-bit recommended). It sets up OpenCV 4 and
# common build tools required by GoCV and the Python inference server.

set -euo pipefail

echo "Starting Raspberry Pi setup (Ubuntu/Debian)..."

echo "Updating apt package index..."
sudo apt-get update -y

echo "Installing build tools and OpenCV development packages..."
sudo apt-get install -y --no-install-recommends \
  build-essential \
  pkg-config \
  cmake \
  git \
  libopencv-dev \
  libgtk-3-dev \
  libavcodec-dev \
  libavformat-dev \
  libswscale-dev \
  libjpeg-dev \
  libpng-dev

echo "Verifying OpenCV 4 via pkg-config..."
if pkg-config --modversion opencv4 >/dev/null 2>&1; then
  echo "OpenCV detected: $(pkg-config --modversion opencv4)"
else
  echo "Error: OpenCV 4 not found by pkg-config. Ensure libopencv-dev is installed."
  exit 1
fi

echo "Checking Go toolchain..."
if command -v go >/dev/null 2>&1; then
  go version
else
  echo "Warning: Go is not installed. Install a recent Go toolchain before building."
fi

echo "Skipping 'uv' installation here; handled by 'make install' (install-uv)."

cat <<'EON'

Setup complete.

Next steps (from the project root):
  make install   # sets up Python deps under py/ with uv
  make build     # builds the Go binary against system OpenCV

Optional checks:
  ./gocamnet --sample-image
  ./gocamnet --list-cameras
  ./gocamnet --show-camera   # requires a desktop/X11 session

If PyTorch wheels fail on aarch64, you can still test the pipeline:
  PYDETECT_MOCK=1 ./gocamnet --show-camera

EON

echo "Raspberry Pi setup finished."


