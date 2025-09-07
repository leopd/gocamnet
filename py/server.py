import argparse
import socket
import struct
import threading
import msgpack
import numpy as np
import torch
from ultralytics import YOLO

HDR_FMT = "<4sIIIIIQQ"
HDR_SIZE = struct.calcsize(HDR_FMT)
FMT_BGR = 0


def load_model(mock: bool):
    if mock:
        return None
    
    # Load YOLO model
    model = YOLO("yolov8n.pt")
    
    # Move to appropriate device
    if torch.backends.mps.is_available():
        model.to("mps")
        device = "mps"
    elif torch.cuda.is_available():
        model.to("cuda")
        device = "cuda"
    else:
        device = "cpu"
    
    # Optimize model for inference
    model.model.eval()  # Set to evaluation mode
    
    # Use torch.compile for PyTorch 2.0+ (significant speedup)
    if hasattr(torch, 'compile') and device != "mps":  # MPS doesn't support compile yet
        try:
            model.model = torch.compile(model.model, mode="reduce-overhead")
            print(f"Model compiled with torch.compile on {device}")
        except Exception as e:
            print(f"torch.compile failed: {e}, using regular model")
    
    # Enable optimizations
    torch.backends.cudnn.benchmark = True if device == "cuda" else False
    torch.backends.cudnn.deterministic = False if device == "cuda" else True
    
    print(f"Model loaded on {device} with optimizations")
    return model


def handle_client(conn, addr, model, score_thresh: float, mock: bool):
    try:
        while True:
            hdr = conn.recv(HDR_SIZE)
            if not hdr or len(hdr) < HDR_SIZE:
                break
            magic, w, h, ch, stride, fmt_code, frame_id, data_len = struct.unpack(HDR_FMT, hdr)
            if magic != b"GCN1" or fmt_code != FMT_BGR or ch != 3:
                break
            buf = b""
            remaining = data_len
            while remaining:
                chunk = conn.recv(remaining)
                if not chunk:
                    break
                buf += chunk
                remaining -= len(chunk)
            if len(buf) != data_len:
                break

            img = np.frombuffer(buf, dtype=np.uint8)
            img = img.reshape((h, stride // ch, ch))[:, :w, :]

            dets = []
            if mock:
                # Return a deterministic single person box for basic plumbing tests
                w2, h2 = int(w // 3), int(h // 3)
                dets.append({"class_id": 0, "class_name": "person", "score": 0.99, "box": [10, 10, 10 + w2, 10 + h2]})
            else:
                # Use torch.no_grad() for inference optimization
                with torch.no_grad():
                    res = model(img, verbose=False)[0]
                names = res.names if hasattr(res, "names") else model.names
                for b in res.boxes:
                    x1, y1, x2, y2 = map(int, b.xyxy[0].tolist())
                    conf = float(b.conf[0].item())
                    cls = int(b.cls[0].item())
                    if conf < score_thresh:
                        continue
                    name = names.get(cls, f"class_{cls}") if isinstance(names, dict) else str(cls)
                    dets.append({"class_id": cls, "class_name": name, "score": conf, "box": [x1, y1, x2, y2]})

            payload = msgpack.packb({"frame_id": int(frame_id), "dets": dets}, use_bin_type=True)
            conn.sendall(struct.pack("<I", len(payload)))
            conn.sendall(payload)
    finally:
        conn.close()


def run_server(host: str, port: int, score_thresh: float, mock: bool):
    model = load_model(mock)
    s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    s.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
    s.bind((host, port))
    s.listen(1)
    print("READY", flush=True)
    try:
        while True:
            conn, addr = s.accept()
            t = threading.Thread(target=handle_client, args=(conn, addr, model, score_thresh, mock), daemon=True)
            t.start()
    finally:
        s.close()


if __name__ == "__main__":
    ap = argparse.ArgumentParser()
    ap.add_argument("--host", default="127.0.0.1")
    ap.add_argument("--port", type=int, default=8081)
    ap.add_argument("--score", type=float, default=0.25)
    ap.add_argument("--mock", action="store_true")
    args = ap.parse_args()
    run_server(args.host, args.port, args.score, args.mock)


