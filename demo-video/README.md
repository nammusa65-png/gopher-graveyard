# Operation demo video (how to use the UI)

`demo.mp4` — 65s, 1280x720, silent, Japanese captions burned in.
A real capture against live servers (Floodgap, Veronica-2, our own
gopherd), recorded because the UI is not self-explanatory at first
glance. Not the Devpost submission video (that one is shot in Phase 2).

Re-shoot: `tools/capture.mjs` (puppeteer-core + chrome-headless-shell,
CDP screencast; needs the gateway on :8070 and gopherd on :70), then
`tools/make_video.py` assembles the frames into mp4 with ffmpeg.
