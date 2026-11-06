// Capture a real operation demo of the Gopher Graveyard UI.
// puppeteer-core + chrome-headless-shell, CDP screencast -> JPEG frames.
// Japanese caption banners and click ripples are injected into the page,
// so the frames themselves carry the explanation.
import puppeteer from "puppeteer-core";
import fs from "node:fs";
import path from "node:path";

const EXE = "/home/hatch/.cache/puppeteer/chrome-headless-shell/linux-156.0.8078.47/chrome-headless-shell-linux64/chrome-headless-shell";
const URL = "http://127.0.0.1:8070/";
const OUT = "/tmp/gg-capture";
const FRAMES = path.join(OUT, "frames");
fs.rmSync(OUT, { recursive: true, force: true });
fs.mkdirSync(FRAMES, { recursive: true });

const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

const browser = await puppeteer.launch({
  executablePath: EXE,
  args: ["--no-sandbox", "--disable-dev-shm-usage", "--hide-scrollbars", "--mute-audio"],
});
const page = await browser.newPage();
await page.setViewport({ width: 1280, height: 720 });

// Search prompt (type-7 items) -> answer "rust".
page.on("dialog", async (d) => { await d.accept("rust"); });

// --- screencast ---
const cdp = await page.createCDPSession();
const frames = []; // {file, ts}
let n = 0;
cdp.on("Page.screencastFrame", async (ev) => {
  const file = `frame_${String(n).padStart(5, "0")}.jpg`;
  fs.writeFileSync(path.join(FRAMES, file), Buffer.from(ev.data, "base64"));
  frames.push({ file, ts: ev.metadata.timestamp });
  n++;
  try { await cdp.send("Page.screencastFrameAck", { sessionId: ev.sessionId }); } catch {}
});
await cdp.send("Page.startScreencast", { format: "jpeg", quality: 70, maxWidth: 1280, maxHeight: 720, everyNthFrame: 1 });

// --- helpers ---
async function setCaption(text) {
  await page.evaluate((t) => {
    let el = document.getElementById("demo-caption");
    if (!el) {
      el = document.createElement("div");
      el.id = "demo-caption";
      el.style.cssText = "position:fixed;left:0;right:0;bottom:0;z-index:99999;background:rgba(10,10,14,.88);color:#fff;font:700 25px 'Noto Sans CJK JP',sans-serif;text-align:center;padding:16px 24px;pointer-events:none;line-height:1.45;";
      document.body.appendChild(el);
    }
    el.textContent = t;
    el.style.display = t ? "block" : "none";
  }, text);
}

async function rippleAndClick(getRect) {
  const rect = await getRect();
  if (!rect) throw new Error("target not found");
  const x = rect.x + rect.w / 2, y = rect.y + rect.h / 2;
  await page.evaluate((x, y) => {
    const r = document.createElement("div");
    r.style.cssText = `position:fixed;left:${x - 6}px;top:${y - 6}px;width:12px;height:12px;border:4px solid #ff5252;border-radius:50%;z-index:99998;pointer-events:none;`;
    document.body.appendChild(r);
    r.animate([
      { transform: "scale(1)", opacity: 1 },
      { transform: "scale(6)", opacity: 0 },
    ], { duration: 550, easing: "ease-out" }).onfinish = () => r.remove();
  }, x, y);
  await sleep(280);
  await page.mouse.click(x, y);
}

// Rect providers: puppeteer serializes evaluate() functions, so closures
// over Node-side variables break. Each provider passes its parameters
// through evaluate's argument channel instead.
const rectOf = (sel) => () => page.evaluate((s) => {
  const el = document.querySelector(s);
  if (!el) return null;
  el.scrollIntoView({ block: "center" });
  const r = el.getBoundingClientRect();
  return { x: r.x, y: r.y, w: r.width, h: r.height };
}, sel);
const rectOfHost = (idx) => () => page.evaluate((i) => {
  const el = document.querySelectorAll("#hosts .host")[i];
  if (!el) return null;
  const r = el.getBoundingClientRect();
  return { x: r.x, y: r.y, w: r.width, h: r.height };
}, idx);
const rectOfItem = (mode, arg) => () => page.evaluate((m, a) => {
  const items = [...document.querySelectorAll("#view a.item")];
  const el = m === "text"
    ? items.find((x) => x.textContent.includes(a))
    : items.find((x) => x.querySelector(".ic")?.textContent === a);
  if (!el) return null;
  el.scrollIntoView({ block: "center" });
  const r = el.getBoundingClientRect();
  return { x: r.x, y: r.y, w: r.width, h: r.height };
}, mode, arg);

const waitItems = (min) => page.waitForFunction(
  (m) => document.querySelectorAll("#view a.item").length >= m, { timeout: 30000 }, min);

// --- scenario ---
await page.goto(URL, { waitUntil: "networkidle2", timeout: 30000 });
await setCaption("Gopher Graveyard 操作デモ ― 死んだはずのGopherをブラウザで掘り起こす");
await sleep(2500);

await setCaption("開くと、墓地マップのホストが生きているか自動チェックされる（●＝生存　✝＝死亡）");
await page.waitForFunction(
  () => [...document.querySelectorAll("#hosts .host")].every((b) => b.classList.contains("alive") || b.classList.contains("dead")),
  { timeout: 45000 });
await sleep(3000);

await setCaption("① ホスト名のボタンを押すと、そのサーバが開く（Floodgap：1999年から続く老舗）");
await rippleAndClick(rectOfHost(1));
await waitItems(5);
await sleep(3000);
await page.evaluate(() => window.scrollTo({ top: 420, behavior: "smooth" }));
await sleep(1800);
await page.evaluate(() => window.scrollTo({ top: 0, behavior: "smooth" }));
await sleep(1200);

await setCaption("② 行を押して辿っていく（📁＝メニュー　📄＝文書　🔍＝検索）");
await rippleAndClick(rectOfItem("icon", "📁"));
await waitItems(3);
await sleep(3000);
await setCaption("左上の ← で戻れる");
await rippleAndClick(rectOf("#back"));
await waitItems(5);
await sleep(2000);

await setCaption("③ アドレスバーに gopher://〜 を打っても開ける（自分で立てたサーバへ）");
await rippleAndClick(rectOf("#addr"));
await page.keyboard.down("Control");
await page.keyboard.press("KeyA");
await page.keyboard.up("Control");
await page.keyboard.type("gopher://localhost", { delay: 60 });
await rippleAndClick(rectOf("#go"));
await waitItems(3);
await sleep(2500);

await setCaption("④ 📄の行はテキスト文書。そのまま読める");
await rippleAndClick(rectOfItem("text", "Why Gopher died"));
await page.waitForFunction(() => document.querySelector("#view pre.text"), { timeout: 20000 });
await sleep(3000);
await page.evaluate(() => window.scrollTo({ top: 380, behavior: "smooth" }));
await sleep(2000);
await rippleAndClick(rectOf("#back"));
await waitItems(3);
await sleep(1500);

await setCaption("⑤ 🔍検索：Veronica-2が全世界のGopherサーバを横断検索する");
await rippleAndClick(rectOfHost(3));
await page.waitForFunction(
  () => [...document.querySelectorAll("#view a.item")].some((a) => a.querySelector(".ic")?.textContent === "🔍"),
  { timeout: 30000 });
await sleep(1500);
await setCaption("検索の行を押して、言葉を入れる（ここでは「rust」）");
await rippleAndClick(rectOfItem("icon", "🔍"));
await waitItems(3);
await sleep(4000);
await page.evaluate(() => window.scrollTo({ top: 350, behavior: "smooth" }));
await sleep(1800);
await page.evaluate(() => window.scrollTo({ top: 0, behavior: "smooth" }));
await sleep(1000);

await setCaption("⑥ Séance mode：生のプロトコル表示。本当にTCPポート70で喋っている証拠が見える");
await rippleAndClick(rectOf(".raw-toggle"));
await rippleAndClick(rectOf("#go"));
await page.waitForFunction(() => { const r = document.getElementById("raw"); return r && !r.hidden && r.textContent.length > 50; }, { timeout: 30000 });
await page.evaluate(() => { const r = document.getElementById("raw"); if (r) window.scrollTo({ top: r.offsetTop - 80, behavior: "smooth" }); });
await sleep(4500);
await page.evaluate(() => window.scrollTo({ top: 0, behavior: "smooth" }));
await sleep(1200);

await setCaption("⑦ 死んでいるホストを開くと、墓石になる（ミネソタ大学：Gopher生誕の地、現在は応答なし）");
await rippleAndClick(rectOfHost(4));
await page.waitForFunction(() => document.querySelector("#view .tombstone"), { timeout: 40000 });
await sleep(4000);

await setCaption("操作はこれだけ：押す・辿る・検索する。あとは世界中の墓地をどうぞ");
await sleep(3500);
await setCaption("");
await sleep(800);

await cdp.send("Page.stopScreencast");
await browser.close();
fs.writeFileSync(path.join(OUT, "frames.json"), JSON.stringify(frames, null, 1));
console.log(`frames: ${frames.length}, span: ${(frames.at(-1).ts - frames[0].ts).toFixed(1)}s`);
