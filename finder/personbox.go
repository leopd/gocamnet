// Package yolo provides person detection (class 0 on COCO) using an ONNX YOLO model.
// No external image libs; uses stdlib + x/image/draw. Inference via onnxruntime_go.
//
// go.mod requires:
//   require (
//     github.com/yalue/onnxruntime_go v0.2.0 // or current
//     golang.org/x/image v0.18.0            // for draw
//   )
package finder

import (
	"errors"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"math"
	"os"
	"sort"

	ort "github.com/yalue/onnxruntime_go"
	xdraw "golang.org/x/image/draw"
)

type Box struct{ X1, Y1, X2, Y2 float32 } // absolute pixels in original image
type Detection struct {
	Box       Box
	Score     float32
	ClassID   int
	ClassName string
}

type Detector struct {
	sess       *ort.DynamicAdvancedSession
	inputName  string
	inputH     int
	inputW     int
	scoreThr   float32
	nmsIoU     float32
	personOnly bool
}

// Options controls Detector construction.
type Options struct {
	// Path to libonnxruntime.dylib compiled with CoreML EP.
	SharedLibPath string
	// ONNX model path (Ultralytics YOLO export without postprocess).
	ModelPath string
	// Input size for letterbox. Defaults to 640x640 if zero.
	InputW, InputH int
	// Score threshold (applied after obj*class prob). Default 0.25.
	ScoreThreshold float32
	// IoU threshold for NMS. Default 0.45.
	NMSIoU float32
	// Whether to keep only "person" (COCO class 0). Default true.
	PersonOnly bool
	// Advanced CoreML EP options. Leave nil for sensible defaults.
	CoreMLOpts map[string]string
	// Explicit input name (Ultralytics default is "images"); leave empty to try common names.
	InputName string
}

// NewDetector creates and warms up an inference session using Core ML EP on Apple Silicon.
func NewDetector(opt Options) (*Detector, error) {
	if opt.SharedLibPath != "" {
		ort.SetSharedLibraryPath(opt.SharedLibPath)
	}
	if err := ort.InitializeEnvironment(ort.WithLogLevelWarning()); err != nil {
		return nil, err
	}

	so, err := ort.NewSessionOptions()
	if err != nil {
		return nil, err
	}

	// Default CoreML EP options: let Core ML choose ANE/GPU/CPU; allow dynamic shapes.
	opts := map[string]string{
		"ModelFormat":              "MLProgram",
		"MLComputeUnits":           "ALL", // CPUAndNeuralEngine|CPUAndGPU|CPUOnly
		"RequireStaticInputShapes": "0",
		"EnableOnSubgraphs":        "0",
	}
	for k, v := range opt.CoreMLOpts {
		opts[k] = v
	}
	if err := so.AppendExecutionProviderCoreMLV2(opts); err != nil {
		so.Destroy()
		return nil, errors.New("CoreML EP not available in this ONNX Runtime build: " + err.Error())
	}
	_ = so.SetGraphOptimizationLevel(ort.GraphOptimizationLevelAll)

	sess, err := ort.NewDynamicAdvancedSession(opt.ModelPath, nil, nil, so)
	so.Destroy()
	if err != nil {
		return nil, err
	}

	inW, inH := opt.InputW, opt.InputH
	if inW == 0 || inH == 0 {
		inW, inH = 640, 640
	}

	score := opt.ScoreThreshold
	if score == 0 {
		score = 0.25
	}
	iou := opt.NMSIoU
	if iou == 0 {
		iou = 0.45
	}

	inputName := opt.InputName
	if inputName == "" {
		// Ultralytics defaults to "images"; some exports use "input" or similar.
		candidates := []string{"images", "input", "images:0"}
		for _, c := range candidates {
			if sess.HasInputName(c) {
				inputName = c
				break
			}
		}
		if inputName == "" {
			// Fallback: pick the first available input
			names := sess.GetInputNames()
			if len(names) == 0 {
				sess.Destroy()
				return nil, errors.New("no inputs found in ONNX model")
			}
			inputName = names[0]
		}
	}

	return &Detector{
		sess:       sess,
		inputName:  inputName,
		inputW:     inW,
		inputH:     inH,
		scoreThr:   score,
		nmsIoU:     iou,
		personOnly: opt.PersonOnly || !opt.PersonOnly, // default true
	}, nil
}

// Close releases native resources.
func (d *Detector) Close() { d.sess.Destroy(); ort.DestroyEnvironment() }

// DetectImagePath runs detection on a JPEG at path.
// Returns best-effort person detections (or all classes if PersonOnly=false).
func (d *Detector) DetectImagePath(path string) ([]Detection, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, err := jpeg.Decode(f)
	if err != nil {
		return nil, err
	}
	return d.Detect(img)
}

// Detect runs detection on an in-memory image.Image.
func (d *Detector) Detect(img image.Image) ([]Detection, error) {
	// 1) Preprocess with letterbox to d.inputW x d.inputH (NCHW float32).
	inTensor, letter, err := d.preprocess(img)
	if err != nil {
		return nil, err
	}
	defer inTensor.Destroy()

	outs, err := d.sess.Run(map[string]ort.Value{d.inputName: inTensor}, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		for _, v := range outs {
			v.Destroy()
		}
	}()

	// Expect exactly one output: [1, N, 85]
	var out ort.Value
	for _, v := range outs {
		out = v
		break
	}
	data, ok := out.(interface{ GetData() []float32 })
	if !ok {
		return nil, errors.New("unexpected output type")
	}
	raw := data.GetData()
	shape := out.GetShape() // [1, N, 85]
	if len(shape) != 3 || shape[0] != 1 || shape[2] < 6 {
		return nil, errors.New("unexpected output shape")
	}
	N := int(shape[1])
	C := int(shape[2]) // 85 on COCO

	// 2) Decode YOLO head: xywh -> xyxy, compute final scores, filter.
	dets := make([]Detection, 0, 64)
	const personClass = 0
	for i := 0; i < N; i++ {
		base := i * C
		cx := raw[base+0]
		cy := raw[base+1]
		w := raw[base+2]
		h := raw[base+3]
		obj := raw[base+4]

		clsID := -1
		clsP := float32(0)
		if d.personOnly {
			clsID = personClass
			clsP = raw[base+5+personClass]
		} else {
			// argmax over classes
			best := float32(0)
			bestIdx := 0
			for c := 0; 5+c < C; c++ {
				p := raw[base+5+c]
				if p > best {
					best = p
					bestIdx = c
				}
			}
			clsID = bestIdx
			clsP = best
		}
		score := obj * clsP
		if score < d.scoreThr {
			continue
		}

		// xywh in letterboxed coord -> xyxy in original pixels
		x1 := cx - w/2
		y1 := cy - h/2
		x2 := cx + w/2
		y2 := cy + h/2
		b := undoLetterbox(x1, y1, x2, y2, letter)

		name := "person"
		if !d.personOnly {
			name = cocoName(clsID)
		}
		dets = append(dets, Detection{
			Box:       b,
			Score:     score,
			ClassID:   clsID,
			ClassName: name,
		})
	}

	// 3) NMS
	return nms(dets, d.nmsIoU), nil
}

// ---------- helpers ----------

// letterbox result to undo later
type letterMeta struct {
	inW, inH int // network size
	imW, imH int // original
	scale    float32
	padX     float32
	padY     float32
}

// preprocess: letterbox to network size, normalize to [0,1], NCHW float32.
func (d *Detector) preprocess(img image.Image) (ort.Value, letterMeta, error) {
	imW := img.Bounds().Dx()
	imH := img.Bounds().Dy()

	// compute scale and pads (same as Ultralytics letterbox default).
	r := float32(min(float64(d.inputW)/float64(imW), float64(d.inputH)/float64(imH)))
	newW := int(math.Round(float64(float32(imW)*r)))
	newH := int(math.Round(float64(float32(imH)*r)))
	padW := d.inputW - newW
	padH := d.inputH - newH
	padL := padW / 2
	padT := padH / 2

	// resize with high quality resampler
	dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
	xdraw.CatmullRom.Scale(dst, dst.Bounds(), toRGBA(img), img.Bounds(), xdraw.Over, nil)

	// place onto letterboxed canvas (gray pad to match common pipelines)
	canvas := image.NewRGBA(image.Rect(0, 0, d.inputW, d.inputH))
	gray := color.RGBA{128, 128, 128, 255}
	draw.Draw(canvas, canvas.Bounds(), &image.Uniform{C: gray}, image.Point{}, draw.Src)
	draw.Draw(canvas, image.Rect(padL, padT, padL+newW, padT+newH), dst, image.Point{}, draw.Over)

	// to NCHW float32 normalized [0,1]
	N := 1
	C := 3
	H := d.inputH
	W := d.inputW
	shape := ort.NewShape(int64(N), int64(C), int64(H), int64(W))
	buf := make([]float32, N*C*H*W)
	// channels
	offR := 0
	offG := H * W
	offB := 2 * H * W
	i := 0
	for y := 0; y < H; y++ {
		for x := 0; x < W; x++ {
			r, g, b, _ := canvas.At(x, y).RGBA()
			buf[offR+i] = float32(r) / 65535.0
			buf[offG+i] = float32(g) / 65535.0
			buf[offB+i] = float32(b) / 65535.0
			i++
		}
	}
	t, err := ort.NewTensor[float32](shape, buf)
	if err != nil {
		return nil, letterMeta{}, err
	}
	return t, letterMeta{
		inW: d.inputW, inH: d.inputH,
		imW: imW, imH: imH,
		scale: r, padX: float32(padL), padY: float32(padT),
	}, nil
}

func toRGBA(img image.Image) *image.RGBA {
	if m, ok := img.(*image.RGBA); ok {
		return m
	}
	r := image.NewRGBA(img.Bounds())
	draw.Draw(r, r.Bounds(), img, img.Bounds().Min, draw.Src)
	return r
}

func undoLetterbox(x1, y1, x2, y2 float32, m letterMeta) Box {
	// remove padding, divide by scale
	x1 = (x1 - m.padX) / m.scale
	y1 = (y1 - m.padY) / m.scale
	x2 = (x2 - m.padX) / m.scale
	y2 = (y2 - m.padY) / m.scale
	// clamp
	x1 = clampf(x1, 0, float32(m.imW-1))
	y1 = clampf(y1, 0, float32(m.imH-1))
	x2 = clampf(x2, 0, float32(m.imW-1))
	y2 = clampf(y2, 0, float32(m.imH-1))
	return Box{x1, y1, x2, y2}
}

func clampf(v, lo, hi float32) float32 {
	if v < lo { return lo }
	if v > hi { return hi }
	return v
}

func iou(a, b Box) float32 {
	ax := max(0, float64(minf(a.X2, b.X2)-maxf(a.X1, b.X1)))
	ay := max(0, float64(minf(a.Y2, b.Y2)-maxf(a.Y1, b.Y1)))
	inter := ax * ay
	aa := float64(a.X2-a.X1) * float64(a.Y2-a.Y1)
	bb := float64(b.X2-b.X1) * float64(b.Y2-b.Y1)
	union := aa + bb - inter
	if union <= 0 { return 0 }
	return float32(inter / union)
}

func nms(dets []Detection, thr float32) []Detection {
	if len(dets) == 0 { return dets }
	sort.Slice(dets, func(i, j int) bool { return dets[i].Score > dets[j].Score })
	keep := make([]Detection, 0, len(dets))
	supp := make([]bool, len(dets))
	for i := 0; i < len(dets); i++ {
		if supp[i] { continue }
		pi := dets[i]
		keep = append(keep, pi)
		for j := i + 1; j < len(dets); j++ {
			if supp[j] { continue }
			// Only suppress boxes of same class to be conservative.
			if dets[j].ClassID != pi.ClassID { continue }
			if iou(pi.Box, dets[j].Box) > thr {
				supp[j] = true
			}
		}
	}
	return keep
}

func cocoName(id int) string {
	// COCO 80 classes; id 0 == person.
	if id == 0 {
		return "person"
	}
	// For brevity, return generic names for non-person classes.
	return "class_" + itoa(id)
}

// simple int->string without strconv to keep imports tight
func itoa(v int) string {
	if v == 0 { return "0" }
	sign := ""
	if v < 0 { sign = "-"; v = -v }
	buf := [20]byte{}
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + (v % 10))
		v /= 10
	}
	return sign + string(buf[i:])
}

func min(a, b float64) float64 { if a < b { return a }; return b }
func max(a, b float64) float64 { if a > b { return a }; return b }
func minf(a, b float32) float32 { if a < b { return a }; return b }
func maxf(a, b float32) float32 { if a > b { return a }; return b }