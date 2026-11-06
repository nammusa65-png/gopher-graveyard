"use strict";

const HOSTS = [
  { name: "Floodgap — community hub since 1999", url: "gopher://gopher.floodgap.com" },
  { name: "SDF — public access UNIX", url: "gopher://sdf.org" },
  { name: "Veronica-2 — search all of Gopherspace", url: "gopher://gopher.floodgap.com/7/v2/vs" },
  { name: "University of Minnesota — where Gopher was born (long dead)", url: "gopher://gopher.umn.edu" },
];

const view = document.getElementById("view");
const rawPre = document.getElementById("raw");
const addr = document.getElementById("addr");
const rawToggle = document.getElementById("rawmode");
let historyStack = [];

const ICONS = { "0": "📄", "1": "📁", "7": "🔍", "9": "💾", "I": "🖼️", "h": "🌐", "i": "  ", "3": "⚠️" };

async function fetchGopher(url, search) {
  view.innerHTML = "<p class='hint'>Digging… contacting the dead protocol over TCP port 70.</p>";
  rawPre.hidden = true;
  let api = "/api/fetch?url=" + encodeURIComponent(url);
  if (search) api += "&search=" + encodeURIComponent(search);
  const res = await fetch(api);
  const data = await res.json();
  if (!res.ok || data.error) {
    view.innerHTML = `<div class="tombstone"><div class="stone">🪦</div>
      <p>Here lies <b>${escapeHTML(url)}</b></p>
      <p class="err">${escapeHTML(data.error || ("HTTP " + res.status))}</p>
      <p>No answer. This host has joined the graveyard.</p></div>`;
    return;
  }
  addr.value = data.url;
  if (rawToggle.checked && data.raw) { rawPre.textContent = data.raw; rawPre.hidden = false; }
  if (data.kind === "text") {
    view.innerHTML = "";
    const pre = document.createElement("pre");
    pre.className = "text";
    pre.textContent = data.text;
    view.appendChild(pre);
    return;
  }
  view.innerHTML = "";
  for (const item of data.items || []) {
    view.appendChild(renderItem(item));
  }
  if (!data.items || data.items.length === 0) {
    view.innerHTML = "<p class='hint'>(empty menu)</p>";
  }
}

function renderItem(item) {
  const icon = ICONS[item.type] ?? "•";
  if (item.type === "i") {
    const p = document.createElement("p");
    p.className = "item info";
    p.textContent = item.display;
    return p;
  }
  if (item.type === "3" || !item.url) {
    const p = document.createElement("p");
    p.className = "item err";
    p.textContent = icon + " " + item.display;
    return p;
  }
  const a = document.createElement("a");
  a.className = "item";
  a.href = "#";
  a.innerHTML = `<span class="ic">${icon}</span>`;
  a.appendChild(document.createTextNode(item.display));
  a.addEventListener("click", (ev) => {
    ev.preventDefault();
    if (item.type === "7") {
      const q = prompt("Search Gopherspace for:");
      if (q) visit(item.url, q);
    } else if (item.type === "h" && item.selector && item.selector.startsWith("URL:")) {
      window.open(item.selector.slice(4), "_blank");
    } else {
      visit(item.url);
    }
  });
  return a;
}

function visit(url, search) {
  historyStack.push(addr.value);
  fetchGopher(url, search).catch((e) => {
    view.innerHTML = `<div class="tombstone"><div class="stone">🪦</div><p class="err">${escapeHTML(String(e))}</p></div>`;
  });
}

function escapeHTML(s) {
  return s.replace(/[&<>"']/g, (c) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c]));
}

document.getElementById("go").addEventListener("click", () => visit(addr.value.trim()));
addr.addEventListener("keydown", (e) => { if (e.key === "Enter") visit(addr.value.trim()); });
document.getElementById("back").addEventListener("click", () => {
  const prev = historyStack.pop();
  if (prev) fetchGopher(prev);
});

// Graveyard map: render known hosts and check which are still alive.
const hostsDiv = document.getElementById("hosts");
for (const h of HOSTS) {
  const b = document.createElement("button");
  b.className = "host";
  b.innerHTML = "<span class='dot'>?</span>" + escapeHTML(h.name);
  b.addEventListener("click", () => visit(h.url));
  hostsDiv.appendChild(b);
  fetch("/api/check?url=" + encodeURIComponent(h.url))
    .then((r) => r.json())
    .then((d) => {
      b.classList.add(d.alive ? "alive" : "dead");
      b.querySelector(".dot").textContent = d.alive ? "●" : "✝";
      b.title = h.url + (d.alive ? " — still alive" : " — no answer (dead)");
    })
    .catch(() => { b.querySelector(".dot").textContent = "?"; });
}
