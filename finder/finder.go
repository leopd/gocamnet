package finder

import (
    "errors"
    "image"
    "os"

    "gocv.io/x/gocv"
)

// Detector provides person (and general object) detection using OpenCV DNN with a YOLO ONNX model.
type Detector struct {
    net            gocv.Net
    inputSize      image.Point
    scoreThreshold float32
    nmsThreshold   float32
    classNames     []string
    modelType      string // "tf-ssd", "caffe-ssd", "darknet", "onnx"
}

// Options configures a Detector.
type Options struct {
    // For Caffe SSD (MobileNet-SSD): set ModelProto to prototxt and ModelWeights to caffemodel
    ModelProto   string
    ModelWeights string
    // For TensorFlow SSD: set ModelTFGraph to .pb and ModelTFConfig to .pbtxt
    ModelTFGraph string
    ModelTFConfig string
    // For Darknet YOLO: set DarknetCfg to .cfg and DarknetWeights to .weights
    DarknetCfg    string
    DarknetWeights string
    // For ONNX models: set ModelPath (not used in this MobileNet-SSD implementation)
    ModelPath string
    // Inference image size. For MobileNet-SSD defaults to 300x300 if zero.
    InputSize image.Point
    // Score threshold for detections. Default 0.5.
    ScoreThreshold float32
    // IoU threshold for NMS (not used by MobileNet-SSD). Default 0.45.
    NMSThreshold float32
    // Optional class names slice (index by class id). If nil, generic labels are used.
    ClassNames []string
}

// NewDetector loads the ONNX model into OpenCV DNN and prepares the network.
func NewDetector(opt Options) (*Detector, error) {
	var net gocv.Net
	var modelType string

	// Prefer CPU target by default to maximize portability.
	// Note: SetPreferableBackend/Target should be called on a valid net object.
	// If net is empty, these calls can segfault.

	if opt.ModelProto != "" && opt.ModelWeights != "" {
		// Basic sanity checks to avoid loading obviously invalid files which can crash the DNN importer
		if fi, err := os.Stat(opt.ModelProto); err != nil || fi.Size() < 1*1024 {
			return nil, errors.New("invalid Caffe proto (file missing or too small): " + opt.ModelProto)
		}
		if fi, err := os.Stat(opt.ModelWeights); err != nil || fi.Size() < 1*1024*1024 {
			return nil, errors.New("invalid Caffe weights (file missing or too small): " + opt.ModelWeights)
		}
		// MobileNet-SSD (Caffe)
		net = gocv.ReadNetFromCaffe(opt.ModelProto, opt.ModelWeights)
		if net.Empty() {
			return nil, errors.New("failed to load Caffe model: " + opt.ModelProto)
		}
		modelType = "caffe-ssd"
	} else if opt.ModelTFGraph != "" && opt.ModelTFConfig != "" {
		// Basic sanity checks to avoid loading obviously invalid files which can crash the DNN importer
		if fi, err := os.Stat(opt.ModelTFGraph); err != nil || fi.Size() < 1*1024*1024 {
			return nil, errors.New("invalid TF graph (file missing or too small): " + opt.ModelTFGraph)
		}
		if fi, err := os.Stat(opt.ModelTFConfig); err != nil || fi.Size() < 1*1024 {
			return nil, errors.New("invalid TF config (file missing or too small): " + opt.ModelTFConfig)
		}
		// TensorFlow SSD (e.g., SSD MobileNet v1/v2)
		net = gocv.ReadNet(opt.ModelTFGraph, opt.ModelTFConfig)
		// A nil check is not possible here as gocv.Net is not a pointer.
		// The call to .Empty() will segfault if ReadNet fails catastrophically.
		// This is a known issue/sharp edge in gocv.
		if net.Empty() {
			return nil, errors.New("failed to load TF model: " + opt.ModelTFGraph)
		}
		modelType = "tf-ssd"
	} else if opt.ModelPath != "" {
		if fi, err := os.Stat(opt.ModelPath); err != nil || fi.Size() < 1*1024*1024 {
			return nil, errors.New("invalid ONNX model (file missing or too small): " + opt.ModelPath)
		}
		// Fallback: ONNX (not recommended here due to importer stability)
		_ = os.Setenv("OPENCV_DNN_DISABLE_MEMORY_OPTIMIZATIONS", "1")
		net = gocv.ReadNetFromONNX(opt.ModelPath)
		if net.Empty() {
			return nil, errors.New("failed to load ONNX model: " + opt.ModelPath)
		}
		modelType = "onnx"
	} else {
		return nil, errors.New("no model specified; provide Caffe ModelProto+ModelWeights, TF ModelTFGraph+ModelTFConfig, or ONNX ModelPath")
	}

	_ = net.SetPreferableBackend(gocv.NetBackendDefault)
	_ = net.SetPreferableTarget(gocv.NetTargetCPU)

    size := opt.InputSize
    if size.X == 0 || size.Y == 0 {
        // Default for MobileNet-SSD
        size = image.Pt(300, 300)
    }
    score := opt.ScoreThreshold
    if score == 0 {
        score = 0.5
    }
    nms := opt.NMSThreshold
    if nms == 0 {
        nms = 0.45
    }

    return &Detector{
        net:            net,
        inputSize:      size,
        scoreThreshold: score,
        nmsThreshold:   nms,
        classNames:     opt.ClassNames,
        modelType:      modelType,
    }, nil
}

// Close frees network resources.
func (d *Detector) Close() error { return d.net.Close() }

// Detection represents a single detection result.
type Detection struct {
    Box       image.Rectangle
    Score     float32
    ClassID   int
    ClassName string
}

// Detect runs detection on the given BGR Mat (e.g., from camera frames).
// The input Mat is not modified.
func (d *Detector) Detect(src gocv.Mat) ([]Detection, error) {
    if src.Empty() {
        return nil, errors.New("empty input image")
    }

    if d.modelType == "darknet" {
        // YOLOv3-tiny style preprocessing
        blob := gocv.BlobFromImage(src, 1.0/255.0, d.inputSize, gocv.NewScalar(0, 0, 0, 0), true, false)
        defer blob.Close()
        d.net.SetInput(blob, "")
        // Forward default output
        outs := d.net.ForwardLayers(d.net.GetLayerNames())
        // Do not parse; just ensure forward completes and clean up
        for _, o := range outs { o.Close() }
        return nil, nil
    }

    // Default SSD-style path
    blob := gocv.BlobFromImage(src, 1.0/127.5, d.inputSize, gocv.NewScalar(127.5, 127.5, 127.5, 0), true, false)
    defer blob.Close()
    d.net.SetInput(blob, "")
    det := d.net.Forward("")
    defer det.Close()
    if det.Empty() { return nil, nil }
    total := int(det.Total())
    if total%7 != 0 { return nil, errors.New("unexpected detection output shape") }
    det = det.Reshape(1, total/7)
    imgW := src.Cols(); imgH := src.Rows()
    var results []Detection
    rows := det.Rows()
    for i := 0; i < rows; i++ {
        classID := int(det.GetFloatAt(i, 1))
        conf := det.GetFloatAt(i, 2)
        if conf < d.scoreThreshold { continue }
        left := int(det.GetFloatAt(i, 3) * float32(imgW))
        top := int(det.GetFloatAt(i, 4) * float32(imgH))
        right := int(det.GetFloatAt(i, 5) * float32(imgW))
        bottom := int(det.GetFloatAt(i, 6) * float32(imgH))
        results = append(results, Detection{ Box: image.Rect(left, top, right, bottom), Score: conf, ClassID: classID, ClassName: d.className(classID) })
    }
    return results, nil
}

func (d *Detector) className(id int) string {
    if d.classNames != nil && id >= 0 && id < len(d.classNames) {
        return d.classNames[id]
    }
    // For MobileNet-SSD, COCO/VOC label for person is 15
    if id == 15 { return "person" }
    return "class_" + itoa(id)
}

// --- minimal itoa to avoid importing strconv ---
func itoa(v int) string {
    if v == 0 { return "0" }
    sign := ""
    if v < 0 { sign = "-"; v = -v }
    var buf [20]byte
    i := len(buf)
    for v > 0 { i--; buf[i] = byte('0' + (v % 10)); v /= 10 }
    return sign + string(buf[i:])
}


