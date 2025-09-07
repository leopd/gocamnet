package finder

import (
    "errors"
    "image"

    "gocv.io/x/gocv"
)

// Detector provides person (and general object) detection using OpenCV DNN with a YOLO ONNX model.
type Detector struct {
    net            gocv.Net
    inputSize      image.Point
    scoreThreshold float32
    nmsThreshold   float32
    classNames     []string
}

// Options configures a Detector.
type Options struct {
    // Path to a YOLO ONNX model (e.g., YOLOv8n/YOLOv5s exported to ONNX).
    ModelPath string
    // Inference image size (typically 640x640 for YOLO). If zero, defaults to 640x640.
    InputSize image.Point
    // Score threshold for detections (post softmax/objectness). Default 0.5.
    ScoreThreshold float32
    // IoU threshold for NMS. Default 0.45.
    NMSThreshold float32
    // Optional class names slice (index by class id). If nil, generic labels are used.
    ClassNames []string
}

// NewDetector loads the ONNX model into OpenCV DNN and prepares the network.
func NewDetector(opt Options) (*Detector, error) {
    if opt.ModelPath == "" {
        return nil, errors.New("ModelPath is required")
    }

    net := gocv.ReadNetFromONNX(opt.ModelPath)
    if net.Empty() {
        return nil, errors.New("failed to load ONNX model: " + opt.ModelPath)
    }

    // Prefer CPU target by default to maximize portability.
    _ = net.SetPreferableBackend(gocv.NetBackendDefault)
    _ = net.SetPreferableTarget(gocv.NetTargetCPU)

    size := opt.InputSize
    if size.X == 0 || size.Y == 0 {
        size = image.Pt(640, 640)
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

    params := gocv.NewImageToBlobParams(
        1.0/255.0,
        d.inputSize,
        gocv.NewScalar(0, 0, 0, 0),
        false, // swapRB: OpenCV uses BGR; most YOLO ONNX expect RGB, but examples set false
        gocv.MatTypeCV32F,
        gocv.DataLayoutNCHW,
        gocv.PaddingModeLetterbox,
        gocv.NewScalar(114.0, 114.0, 114.0, 0),
    )
    blob := gocv.BlobFromImageWithParams(src, params)
    defer blob.Close()

    d.net.SetInput(blob, "")

    outputNames := getOutputNames(&d.net)
    if len(outputNames) == 0 {
        return nil, errors.New("failed to get output layer names")
    }

    outs := d.net.ForwardLayers(outputNames)
    defer func() {
        for _, o := range outs { o.Close() }
    }()

    boxes, confidences, classIDs := performDetection(outs)
    if len(boxes) == 0 {
        return nil, nil
    }

    // Map from blob rects to image rects
    iboxes := params.BlobRectsToImageRects(boxes, image.Pt(src.Cols(), src.Rows()))
    indices := gocv.NMSBoxes(iboxes, confidences, d.scoreThreshold, d.nmsThreshold)

    results := make([]Detection, 0, len(indices))
    for _, idx := range indices {
        cid := classIDs[idx]
        name := d.className(cid)
        results = append(results, Detection{
            Box:       iboxes[idx],
            Score:     confidences[idx],
            ClassID:   cid,
            ClassName: name,
        })
    }
    return results, nil
}

func (d *Detector) className(id int) string {
    if d.classNames != nil && id >= 0 && id < len(d.classNames) {
        return d.classNames[id]
    }
    if id == 0 { return "person" }
    return "class_" + itoa(id)
}

// getOutputNames mirrors the gocv example to fetch YOLO heads.
func getOutputNames(net *gocv.Net) []string {
    var names []string
    for _, i := range net.GetUnconnectedOutLayers() {
        layer := net.GetLayer(i)
        n := layer.GetName()
        if n != "_input" {
            names = append(names, n)
        }
    }
    return names
}

// performDetection decodes YOLO output tensors to bounding boxes, confidences and class IDs.
func performDetection(outs []gocv.Mat) ([]image.Rectangle, []float32, []int) {
    var boxes []image.Rectangle
    var confidences []float32
    var classIDs []int

    if len(outs) == 0 {
        return boxes, confidences, classIDs
    }

    // YOLOv8 requires transpose [1,84,N] -> [1,N,84] like in gocv example
    gocv.TransposeND(outs[0], []int{0, 2, 1}, &outs[0])

    for _, out := range outs {
        out = out.Reshape(1, out.Size()[1])
        rows := out.Rows()
        cols := out.Cols()
        for i := 0; i < rows; i++ {
            row := out.RowRange(i, i+1)
            scores := row.ColRange(4, cols)
            _, conf, _, classPt := gocv.MinMaxLoc(scores)
            if conf > 0.5 { // pre-filter before NMS
                cx := out.GetFloatAt(i, 0)
                cy := out.GetFloatAt(i, 1)
                w := out.GetFloatAt(i, 2)
                h := out.GetFloatAt(i, 3)
                left := cx - w/2
                top := cy - h/2
                right := cx + w/2
                bottom := cy + h/2
                boxes = append(boxes, image.Rect(int(left), int(top), int(right), int(bottom)))
                confidences = append(confidences, float32(conf))
                classIDs = append(classIDs, classPt.X)
            }
        }
    }
    return boxes, confidences, classIDs
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


