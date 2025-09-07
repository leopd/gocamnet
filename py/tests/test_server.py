import os
import socket
import struct
import subprocess
import sys
import time
import msgpack

HDR_FMT = "<4sIIIIIQQ"
FMT_BGR = 0


def get_free_port():
    s = socket.socket()
    s.bind(("127.0.0.1", 0))
    port = s.getsockname()[1]
    s.close()
    return port


def start_server(port):
    # Run using the same interpreter as pytest (already inside uv's venv)
    cmd = [sys.executable, "server.py", "--host", "127.0.0.1", "--port", str(port), "--mock"]
    p = subprocess.Popen(cmd, cwd=os.path.join(os.path.dirname(__file__), ".."), stdout=subprocess.PIPE, stderr=subprocess.STDOUT, text=True)
    # wait for READY line
    deadline = time.time() + 15
    lines = []
    while time.time() < deadline:
        line = p.stdout.readline()
        if not line:
            break
        lines.append(line)
        if line.strip() == "READY":
            return p
    raise RuntimeError("server failed to start. Output: " + "".join(lines))


def test_mock_roundtrip():
    port = get_free_port()
    p = start_server(port)
    try:
        s = socket.create_connection(("127.0.0.1", port), timeout=5)
        w, h, ch = 32, 24, 3
        stride = w * ch
        data = bytes(w * h * ch)
        hdr = struct.pack(HDR_FMT, b"GCN1", w, h, ch, stride, FMT_BGR, 1, len(data))
        s.sendall(hdr)
        s.sendall(data)
        # read msgpack length
        ln = s.recv(4)
        plen = struct.unpack("<I", ln)[0]
        payload = b""
        while len(payload) < plen:
            payload += s.recv(plen - len(payload))
        resp = msgpack.unpackb(payload)
        assert "dets" in resp and len(resp["dets"]) >= 1
    finally:
        p.kill()


