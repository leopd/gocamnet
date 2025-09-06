#!/bin/bash

# This script installs OpenCV 4 on macOS using Homebrew.

# Exit immediately if a command exits with a non-zero status.
set -e

echo "Starting OpenCV 4 installation on macOS..."

# 1. Ensure Homebrew is installed
if ! command -v brew &> /dev/null
then
    echo "Homebrew not found. Installing Homebrew..."
    /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
else
    echo "Homebrew is already installed."
fi

# 2. Update Homebrew and upgrade any existing packages
echo "Updating Homebrew and upgrading packages..."
brew update
brew upgrade

# 3. Install OpenCV 4
echo "Installing OpenCV 4... (This might take some time)"
brew install opencv

# 4. Verify OpenCV installation (optional, but good practice)
if pkg-config --modversion opencv4 &> /dev/null
then
    echo "OpenCV 4 installed successfully! Version: $(pkg-config --modversion opencv4)"
else
    echo "Error: OpenCV 4 installation verification failed."
    exit 1
fi

# Set PKG_CONFIG_PATH for GoCV
# Homebrew installs pkgconfig files to /usr/local/opt/opencv/lib/pkgconfig
export PKG_CONFIG_PATH="$(brew --prefix opencv)/lib/pkgconfig:$PKG_CONFIG_PATH"
echo "PKG_CONFIG_PATH set to: $PKG_CONFIG_PATH"

echo "OpenCV 4 setup complete!"
