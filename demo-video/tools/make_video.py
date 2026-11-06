#!/usr/bin/env python3
"""Assemble screencast frames (VFR, timestamped) into a 30fps mp4."""
import json
import subprocess
import sys

CAP = "/tmp/gg-capture"
frames = json.load(open(f"{CAP}/frames.json"))
frames.sort(key=lambda f: f["ts"])
t0 = frames[0]["ts"]

lines = []
for i, fr in enumerate(frames):
    if i + 1 < len(frames):
        dur = frames[i + 1]["ts"] - fr["ts"]
    else:
        dur = 1.5
    dur = min(max(dur, 0.02), 5.0)
    lines.append(f"file 'frames/{fr['file']}'")
    lines.append(f"duration {dur:.3f}")
# concat demuxer needs the last file repeated to honor its duration
lines.append(f"file 'frames/{frames[-1]['file']}'")
open(f"{CAP}/concat.txt", "w").write("\n".join(lines) + "\n")

out = sys.argv[1] if len(sys.argv) > 1 else "/home/hatch/workspace/gopher-graveyard/demo-video/demo.mp4"
cmd = [
    "ffmpeg", "-y", "-f", "concat", "-safe", "0", "-i", f"{CAP}/concat.txt",
    "-vf", "fps=30,scale=1280:720:flags=lanczos,format=yuv420p",
    "-c:v", "libx264", "-crf", "20", "-preset", "medium",
    "-movflags", "+faststart", out,
]
print("+", " ".join(cmd))
subprocess.run(cmd, check=True)
print("wrote", out)
