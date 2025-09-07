NAME := gocamnet
OS := $(shell uname -s)
ARCH := $(shell uname -m)

# Model download location (not checked into source control)
MODEL_DIR := models
EXECUTABLE := gocamnet

# Model Definitions (User must manually download these)
# Caffe MobileNet-SSD
MODEL_PROTO := MobileNetSSD_deploy.prototxt
MODEL_WEIGHTS := MobileNetSSD_deploy.caffemodel


# Darknet YOLOv3-tiny
DARKNET_CFG := yolov3-tiny.cfg
DARKNET_WEIGHTS := yolov3-tiny.weights
DARKNET_URL_CFG := https://github.com/opencv/opencv_3rdparty/raw/dnn_samples/dnn/$(DARKNET_CFG)
DARKNET_URL_WEIGHTS := https://pjreddie.com/media/files/$(DARKNET_WEIGHTS)
DARKNET_URL_WEIGHTS_ALT := https://github.com/AlexeyAB/darknet/releases/download/darknet_yolo_v3_optimal/$(DARKNET_WEIGHTS)

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

.PHONY: all build run run-sample-image run-camera list-cameras test clean install model download-model

all: build

build: model
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

test: model
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

# Download the model if not present
model: download-model

.PHONY: download-model
download-model:
	@mkdir -p $(MODEL_DIR)
	@echo "Model download is a manual step. See README.md for instructions."

clean-models:
	rm -f $(MODEL_DIR)/*.pb $(MODEL_DIR)/*.pbtxt $(MODEL_DIR)/*.onnx
	rm -f $(MODEL_DIR)/*.cfg $(MODEL_DIR)/*.weights $(MODEL_DIR)/*.caffemodel $(MODEL_DIR)/*.prototxt
