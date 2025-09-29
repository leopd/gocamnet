package pydetect

import (
	"bufio"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/vmihailenco/msgpack/v5"
	"gocv.io/x/gocv"
)

type Detector struct {
	cmd     *exec.Cmd
	conn    net.Conn
	score   float32
	cancel  context.CancelFunc
	readyCh chan struct{}
}

type Detection struct {
	Box       image.Rectangle
	Score     float32
	ClassID   int
	ClassName string
}

func New(ctx context.Context, projectRoot string, score float32) (*Detector, error) {
	if score <= 0 {
		score = 0.25
	}
	dctx, cancel := context.WithCancel(ctx)
	d := &Detector{
		score:   score,
		cancel:  cancel,
		readyCh: make(chan struct{}),
	}

	pyDir := filepath.Join(projectRoot, "py")
	server := filepath.Join(pyDir, "server.py")

	args := []string{"run", server, "--host", "127.0.0.1", "--port", "0", "--score", fmt.Sprintf("%.3f", d.score)}
	if os.Getenv("PYDETECT_MOCK") == "1" {
		args = append(args, "--mock")
	}
	d.cmd = exec.CommandContext(dctx, "uv", args...)
	d.cmd.Dir = pyDir
	d.cmd.Env = append(os.Environ(), "PYTHONUNBUFFERED=1")

	stdout, err := d.cmd.StdoutPipe()
	if err != nil {
		cancel()
		return nil, err
	}
	stderr, err := d.cmd.StderrPipe()
	if err != nil {
		cancel()
		return nil, err
	}

	if err := d.cmd.Start(); err != nil {
		cancel()
		return nil, err
	}

	readyPortCh := make(chan int, 1)
	readyErrCh := make(chan error, 1)

	// Watch for READY and capture actual port
	go func() {
		sc := bufio.NewScanner(stdout)
		for sc.Scan() {
			line := sc.Text()
			if strings.HasPrefix(line, "READY") {
				parts := strings.Fields(line)
				if len(parts) < 2 {
					readyErrCh <- errors.New("python server did not provide port")
					return
				}
				value := parts[1]
				port, err := strconv.Atoi(value)
				if err != nil || port <= 0 {
					readyErrCh <- fmt.Errorf("invalid READY port %q", value)
					return
				}
				readyPortCh <- port
				close(d.readyCh)
				return
			}
			fmt.Fprintln(os.Stderr, line)
		}
		if err := sc.Err(); err != nil {
			readyErrCh <- err
		} else {
			readyErrCh <- errors.New("python server exited before READY")
		}
	}()

	// Surface server stderr to ours
	go func() { _, _ = io.Copy(os.Stderr, stderr) }()

	var port int
	select {
	case port = <-readyPortCh:
	case err := <-readyErrCh:
		d.Close()
		return nil, err
	case <-time.After(15 * time.Second):
		d.Close()
		return nil, errors.New("python server failed to become ready")
	}

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		d.Close()
		return nil, fmt.Errorf("connect failed: %w", err)
	}
	d.conn = conn
	return d, nil
}

func (d *Detector) Close() error {
	if d.conn != nil {
		_ = d.conn.Close()
	}
	if d.cancel != nil {
		d.cancel()
	}
	shutdownProcess(d.cmd)
	return nil
}

var hdrMagic = [4]byte{'G', 'C', 'N', '1'}

const fmtBGR uint32 = 0

// header: "GCN1"(4) + w(u32) h(u32) ch(u32) stride(u32) fmt(u32) frame_id(u64) data_len(u64)
func (d *Detector) Detect(mat gocv.Mat, frameID uint64) ([]Detection, error) {
	if mat.Empty() {
		return nil, errors.New("empty mat")
	}
	w := uint32(mat.Cols())
	h := uint32(mat.Rows())
	ch := uint32(mat.Channels())
	if ch != 3 {
		return nil, fmt.Errorf("expected 3 channels, got %d", ch)
	}
	data, err := mat.DataPtrUint8()
	if err != nil || len(data) == 0 {
		return nil, errors.New("failed to access mat bytes")
	}
	stride := uint32(w * ch)
	dataLen := uint64(len(data))

	hdr := make([]byte, 4+4*5+8+8)
	copy(hdr[:4], hdrMagic[:])
	off := 4
	putU32 := func(v uint32) { binary.LittleEndian.PutUint32(hdr[off:], v); off += 4 }
	putU64 := func(v uint64) { binary.LittleEndian.PutUint64(hdr[off:], v); off += 8 }

	putU32(w)
	putU32(h)
	putU32(ch)
	putU32(stride)
	putU32(fmtBGR)
	putU64(frameID)
	putU64(dataLen)

	if _, err := d.conn.Write(hdr); err != nil {
		return nil, err
	}
	if _, err := d.conn.Write(data); err != nil {
		return nil, err
	}

	lenBuf := make([]byte, 4)
	if _, err := io.ReadFull(d.conn, lenBuf); err != nil {
		return nil, err
	}
	plLen := binary.LittleEndian.Uint32(lenBuf)
	pl := make([]byte, plLen)
	if _, err := io.ReadFull(d.conn, pl); err != nil {
		return nil, err
	}

	var resp struct {
		FrameID uint64 `msgpack:"frame_id"`
		Dets    []struct {
			ClassID   int     `msgpack:"class_id"`
			ClassName string  `msgpack:"class_name"`
			Score     float64 `msgpack:"score"`
			Box       [4]int  `msgpack:"box"`
		} `msgpack:"dets"`
	}
	if err := msgpack.Unmarshal(pl, &resp); err != nil {
		return nil, err
	}
	out := make([]Detection, 0, len(resp.Dets))
	for _, d2 := range resp.Dets {
		out = append(out, Detection{
			Box:       image.Rect(d2.Box[0], d2.Box[1], d2.Box[2], d2.Box[3]),
			Score:     float32(d2.Score),
			ClassID:   d2.ClassID,
			ClassName: d2.ClassName,
		})
	}
	return out, nil
}

func shutdownProcess(cmd *exec.Cmd) {
	if cmd == nil {
		return
	}
	if cmd.ProcessState != nil {
		return
	}

	done := make(chan struct{})
	go func() {
		if err := cmd.Wait(); err != nil && !errors.Is(err, os.ErrProcessDone) {
			// Process exited with an error; nothing further to do here.
		}
		close(done)
	}()

	if cmd.Process != nil {
		if err := cmd.Process.Signal(os.Interrupt); err != nil {
			_ = cmd.Process.Kill()
		}
	}

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		<-done
	}
}
