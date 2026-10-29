// previewer is the Alpine component behind the previewer page. The page is
// rendered by PopUI; this filters the examples, renders the selected one
// through the WebAssembly build onto a true-ratio sheet, and benchmarks WASM
// against the server's native render time.
function previewer() {
  const PAGES = { A4: [210, 297], Letter: [215.9, 279.4], DIN5008: [210, 297] };
  // DIN 5008 forms A and B: envelope window (x, y, w, h), fold marks and the
  // punch mark, in mm from the top-left corner of the sheet.
  const DIN = {
    A: { window: [20, 27, 90, 45], folds: [87, 192], punch: 148.5 },
    B: { window: [20, 45, 90, 45], folds: [105, 210], punch: 148.5 },
  };
  const GROUPS = {
    country: (e) => (e.country ? [e.country] : []),
    kind: (e) => [e.kind],
    addon: (e) => e.addons ?? [],
    tag: (e) => e.tags ?? [],
  };
  const PER_EXAMPLE = ["locale", "void", "sandbox", "adjustment_mode"];
  const median = (xs) => {
    const s = [...xs].sort((a, b) => a - b);
    return s.length ? s[Math.floor(s.length / 2)] : 0;
  };
  const fmt = (ms) => ms.toFixed(2);
  const sources = {};
  let mm = 0;

  return {
    entries: [],
    filters: { country: [], kind: [], addon: [], tag: [] },
    q: "",
    selected: null,
    params: { locale: "", layout: "", label: "", watermark: "", date_format: "", void: false, sandbox: false, adjustment_mode: false },
    window: "",
    zoom: "fit",
    margins: false,
    rounds: 3,
    ready: false,
    running: false,
    bench: null,
    meta: "",
    error: "",
    pages: 0,
    lastShown: null,

    // ---- Boot --------------------------------------------------------
    async init() {
      this.entries = JSON.parse(document.getElementById("catalog").textContent);
      const probe = document.createElement("div");
      probe.style.cssText = "position:absolute;visibility:hidden;width:100mm;height:0";
      document.body.appendChild(probe);
      mm = probe.getBoundingClientRect().width / 100;
      probe.remove();

      this.readURL();
      this.ensureSelected();
      this.$watch("q", () => this.filtersChanged());
      this.$watch("params", () => this.show());
      new ResizeObserver(() => this.fitPaper()).observe(this.$refs.canvas);
      document.addEventListener("keydown", (ev) => this.keydown(ev));
      this.show();

      try {
        const t = performance.now();
        const go = new Go();
        const { instance } = await WebAssembly.instantiateStreaming(fetch("/wasm/gobl.html.wasm"), go.importObject);
        const up = new Promise((r) => addEventListener("gobl:ready", r, { once: true }));
        go.run(instance);
        await up;
        this.ready = true;
        const v = goblVersion();
        this.version = `gobl.html ${v.version} · gobl ${v.gobl}`;
        console.log(`[previewer] WASM ready in ${(performance.now() - t).toFixed(0)} ms · ${this.version}`);
        this.show();
      } catch (err) {
        this.error = `WASM failed to load: ${err}`;
      }
    },

    // ---- URL state ---------------------------------------------------
    readURL() {
      const u = new URLSearchParams(location.search);
      this.q = u.get("q") ?? "";
      for (const k of Object.keys(GROUPS)) this.filters[k] = (u.get(k) ?? "").split(",").filter(Boolean);
      this.selected = u.get("file");
    },
    writeURL() {
      const u = new URLSearchParams();
      if (this.q) u.set("q", this.q);
      for (const k of Object.keys(GROUPS)) if (this.filters[k].length) u.set(k, this.filters[k].join(","));
      if (this.selected) u.set("file", this.selected);
      const qs = u.toString();
      history.replaceState(null, "", qs ? "?" + qs : location.pathname);
    },

    // ---- Filtering ---------------------------------------------------
    matchesSearch(e) {
      const q = this.q.trim().toLowerCase();
      if (!q) return true;
      const hay = [e.file, e.title, e.supplier, e.kind, e.country, ...(e.tags ?? []), ...(e.addons ?? [])].join(" ").toLowerCase();
      return q.split(/\s+/).every((w) => hay.includes(w));
    },
    matchesGroup(e, key) {
      const sel = this.filters[key];
      return sel.length === 0 || GROUPS[key](e).some((v) => sel.includes(v));
    },
    matches(e, except) {
      return this.matchesSearch(e) && Object.keys(GROUPS).every((k) => k === except || this.matchesGroup(e, k));
    },
    get visible() {
      return this.entries.filter((e) => this.matches(e));
    },
    isVisible(file) {
      const e = this.entries.find((x) => x.file === file);
      return !!e && this.matches(e);
    },
    // Facet counts honour the other groups and the search, so they show
    // what each click would leave.
    count(key, value) {
      let n = 0;
      for (const e of this.entries) if (this.matches(e, key) && GROUPS[key](e).includes(value)) n++;
      return n;
    },
    isOn(key, value) {
      return this.filters[key].includes(value);
    },
    toggle(key, value) {
      const i = this.filters[key].indexOf(value);
      if (i >= 0) this.filters[key].splice(i, 1);
      else this.filters[key].push(value);
      this.filtersChanged();
    },
    hasFilters() {
      return !!this.q || Object.values(this.filters).some((f) => f.length);
    },
    clearAll() {
      this.q = "";
      for (const k of Object.keys(GROUPS)) this.filters[k] = [];
      this.filtersChanged();
    },
    filtersChanged() {
      this.ensureSelected();
      this.writeURL();
      this.show();
    },
    ensureSelected() {
      const vis = this.visible;
      if (this.selected && !vis.some((e) => e.file === this.selected)) this.selected = null;
      if (!this.selected && vis.length) this.selected = vis[0].file;
    },
    countLabel() {
      const n = this.visible.length;
      return n === this.entries.length ? `${n} examples` : `${n} of ${this.entries.length}`;
    },

    // ---- Selection ---------------------------------------------------
    get current() {
      return this.entries.find((e) => e.file === this.selected) ?? null;
    },
    currentTitle() {
      if (this.bench) return "Benchmark";
      const e = this.current;
      return e ? e.title.charAt(0).toUpperCase() + e.title.slice(1) : "—";
    },
    select(file) {
      this.selected = file;
      this.writeURL();
      this.$nextTick(() => document.querySelector(`[data-row][data-file="${file}"]`)?.scrollIntoView({ block: "nearest" }));
      this.show();
    },
    move(delta) {
      const vis = this.visible;
      if (!vis.length) return;
      const i = vis.findIndex((e) => e.file === this.selected);
      const j = Math.min(vis.length - 1, Math.max(0, (i < 0 ? 0 : i) + delta));
      this.select(vis[j].file);
    },
    keydown(ev) {
      const typing = ["INPUT", "SELECT", "TEXTAREA"].includes(ev.target.tagName);
      if (ev.key === "/" && !typing) {
        ev.preventDefault();
        this.$refs.q.focus();
        this.$refs.q.select();
      } else if ((ev.key === "ArrowDown" || ev.key === "ArrowUp") && (!typing || ev.target === this.$refs.q)) {
        ev.preventDefault();
        this.move(ev.key === "ArrowDown" ? 1 : -1);
      } else if (ev.key === "Escape" && ev.target === this.$refs.q) {
        this.q = "";
      }
    },

    // ---- Rendering ---------------------------------------------------
    entryParams(e) {
      const p = {};
      if (e.locale) p.locale = e.locale;
      if (e.void) p.void = true;
      if (e.sandbox) p.sandbox = true;
      if (e.adjustment_mode) p.adjustment_mode = true;
      return p;
    },
    requestParams() {
      const p = {};
      for (const [k, v] of Object.entries(this.params)) if (v) p[k] = v;
      return p;
    },
    serverURL(file, ext, p = this.requestParams()) {
      return "/" + file.replace(/\.json$/, "." + ext) + "?" + new URLSearchParams(p);
    },
    async show() {
      if (this.bench) return; // the benchmark owns the canvas until closed
      const e = this.current;
      const frame = this.$refs.frame;
      if (!e) {
        this.meta = "";
        frame.removeAttribute("srcdoc");
        frame.src = "about:blank";
        return;
      }
      if (e.file !== this.lastShown) {
        // The example's own state replaces the per-example parameters; the
        // persistent ones (layout, label, …) carry over.
        const p = this.entryParams(e);
        for (const k of PER_EXAMPLE) this.params[k] = p[k] ?? (typeof this.params[k] === "boolean" ? false : "");
        this.lastShown = e.file;
      }
      const p = this.requestParams();
      if (!this.ready) {
        // Until the WASM is up the server renders, so browsing isn't blocked.
        frame.removeAttribute("srcdoc");
        frame.src = this.serverURL(e.file, "html", p);
        this.meta = "server render";
        return;
      }
      if (!sources[e.file]) sources[e.file] = await (await fetch("/api/examples/" + e.file)).text();
      if (this.selected !== e.file) return; // moved on while fetching
      const t = performance.now();
      const res = goblRender(sources[e.file], p);
      const ms = performance.now() - t;
      if (res.error) {
        this.error = res.error;
        this.meta = "";
        frame.srcdoc = "";
        return;
      }
      this.error = "";
      this.meta = `wasm ${ms.toFixed(1)} ms · ${(res.html.length / 1024).toFixed(0)} KB`;
      frame.srcdoc = res.html;
    },
    // The envelope window follows the layout: DIN 5008 turns Form B on,
    // any other layout turns the window off. It can still be set by hand
    // afterwards in the options.
    layoutChanged() {
      this.window = this.params.layout === "DIN5008" ? this.window || "B" : "";
      this.fitPaper();
    },

    // ---- Paper -------------------------------------------------------
    // Each iframe is laid out at the real page width, so line breaks match
    // print, then scaled to the column. Its height grows to whole pages and
    // page boundaries are drawn where the page would end. While a benchmark
    // runs the browser and server sheets share the width.
    fitPaper() {
      const r = this.$refs;
      if (!r.frame || !mm) return;
      const [wmm, hmm] = PAGES[this.params.layout || "A4"] ?? PAGES.A4;
      const avail = r.canvas.clientWidth - 56;
      const maxW = this.running ? (avail - 28) / 2 : avail;
      this.pages = this.fitSheet(r.frame, r.sheet, r.paper, r.overlay, wmm, hmm, maxW, !this.bench);
      if (this.running) this.fitSheet(r.frame2, r.sheet2, r.paper2, null, wmm, hmm, maxW, false);
      const text = this.meta.replace(/ · \d+ pages?$/, "");
      this.meta = this.pages > 1 && !this.running ? `${text} · ${this.pages} pages` : text;
    },
    fitSheet(frame, sheet, paper, overlay, wmm, hmm, maxW, decorate) {
      const pageW = wmm * mm;
      const pageH = hmm * mm;
      // Measure at one page tall first; the document's min-height would
      // otherwise echo back whatever height was set last.
      frame.style.width = pageW + "px";
      frame.style.height = pageH + "px";
      let contentH = pageH;
      try {
        const d = frame.contentDocument;
        if (d && d.documentElement) contentH = Math.max(d.documentElement.scrollHeight, d.body?.scrollHeight ?? 0);
      } catch (_) {}
      const pages = Math.max(1, Math.ceil((contentH - 1) / pageH));
      const sheetH = pages * pageH;
      frame.style.height = sheetH + "px";

      const s = this.zoom === "fit" || this.running ? Math.min(1, maxW / pageW) : Number(this.zoom);
      sheet.style.width = pageW + "px";
      sheet.style.height = sheetH + "px";
      sheet.style.transform = `scale(${s})`;
      paper.style.width = pageW * s + "px";
      paper.style.height = sheetH * s + "px";

      if (!overlay) return pages;
      let html = "";
      for (let i = 1; i < pages; i++) html += `<div class="pv-break" style="top:${i * hmm}mm"><span>page ${i + 1}</span></div>`;
      if (decorate && this.margins) html += `<div class="pv-edge"></div>`;
      const din = decorate ? DIN[this.window] : null;
      if (din) {
        const [x, y, w, h] = din.window;
        html += `<div class="pv-window" style="left:${x}mm;top:${y}mm;width:${w}mm;height:${h}mm"><span>DIN 5008 ${this.window} window ${w}×${h} mm</span></div>`;
        for (const f of din.folds) html += `<div class="pv-fold" style="top:${f}mm"></div>`;
        html += `<div class="pv-fold punch" style="top:${din.punch}mm"></div>`;
      }
      overlay.innerHTML = html;
      return pages;
    },

    // ---- Benchmark ---------------------------------------------------
    // Renders the listed examples through WASM in the browser and natively
    // in the server for several rounds and compares. The native time comes
    // from the X-Render-Ms header, so it excludes parsing and transfer. Both
    // renders are shown side by side as they happen, from the same bytes the
    // timings were taken on.
    async runBenchmark() {
      if (this.running) return;
      this.running = true;
      this.bench = { running: true };
      this.error = "";
      const r = this.$refs;
      let lastPaint = 0;
      const paint = (wasmHTML, nativeHTML, force) => {
        const now = performance.now();
        // Painting two sheets costs layout time on the same thread, so keep
        // it to a few times a second; the timings stay representative.
        if (!force && now - lastPaint < 400) return;
        lastPaint = now;
        r.frame.srcdoc = wasmHTML;
        // The server's links are relative; a base makes them resolve inside srcdoc.
        r.frame2.srcdoc = nativeHTML.replace(/<head>/i, `<head><base href="${location.origin}/">`);
      };
      try {
        const rounds = Math.max(1, Number(this.rounds) || 1);
        const rows = this.visible.map((e) => ({ e, wasm: [], native: [], trip: [], bytes: 0 }));
        if (!rows.length) throw new Error("nothing to benchmark");
        for (const row of rows) {
          if (!sources[row.e.file]) sources[row.e.file] = await (await fetch("/api/examples/" + row.e.file)).text();
        }
        this.$nextTick(() => this.fitPaper());
        // Warm up both sides once so JIT and caches don't skew round one.
        for (const row of rows) {
          this.meta = `warming up: ${row.e.file}`;
          const res = goblRender(sources[row.e.file], this.entryParams(row.e));
          const html = await fetch(this.serverURL(row.e.file, "html", this.entryParams(row.e))).then((x) => x.text());
          paint(res.html ?? "", html, false);
        }
        for (let round = 1; round <= rounds; round++) {
          for (const row of rows) {
            this.meta = `round ${round}/${rounds} · ${row.e.file}`;
            const p = this.entryParams(row.e);
            const t0 = performance.now();
            const res = goblRender(sources[row.e.file], p);
            row.wasm.push(performance.now() - t0);
            if (res.error) throw new Error(`${row.e.file}: ${res.error}`);
            row.bytes = res.html.length;
            const t1 = performance.now();
            const x = await fetch(this.serverURL(row.e.file, "html", p));
            const html = await x.text();
            row.trip.push(performance.now() - t1);
            row.native.push(Number(x.headers.get("X-Render-Ms")) || NaN);
            paint(res.html, html, round === rounds && row === rows[rows.length - 1]);
          }
        }
        this.running = false;
        this.bench = this.summarise(rounds, rows);
        this.meta = `${rows.length} examples × ${rounds} rounds · WASM ${this.bench.total.wasmMed} ms · native ${this.bench.total.nativeMed} ms · ${this.bench.total.ratio}`;
        r.frame.srcdoc = await this.reportHTML(this.bench);
      } catch (err) {
        this.error = String(err);
        this.running = false;
        this.bench = null;
        this.show();
      }
    },
    closeBenchmark() {
      this.bench = null;
      this.running = false;
      this.show();
    },
    summarise(rounds, rows) {
      const all = (k) => rows.flatMap((r) => r[k]);
      const sum = (xs) => xs.reduce((a, b) => a + b, 0);
      const ratio = median(all("wasm")) / median(all("native"));
      return {
        running: false,
        rounds,
        date: new Date().toISOString().slice(0, 10),
        browser: navigator.userAgent.match(/(Chrome|Safari|Firefox)\/[\d.]+/)?.[0] ?? navigator.userAgent,
        rows: rows.map((r) => ({
          e: r.e,
          kb: (r.bytes / 1024).toFixed(0),
          wasmMed: fmt(median(r.wasm)),
          wasmMin: fmt(Math.min(...r.wasm)),
          nativeMed: fmt(median(r.native)),
          nativeMin: fmt(Math.min(...r.native)),
          ratio: (median(r.wasm) / median(r.native)).toFixed(1) + "×",
          trip: fmt(median(r.trip)),
        })),
        total: {
          wasmMed: fmt(median(all("wasm"))),
          wasmMin: fmt(Math.min(...all("wasm"))),
          wasmSum: fmt(sum(all("wasm"))),
          nativeMed: fmt(median(all("native"))),
          nativeMin: fmt(Math.min(...all("native"))),
          nativeSum: fmt(sum(all("native"))),
          trip: fmt(median(all("trip"))),
          ratio: ratio.toFixed(1) + "×",
        },
      };
    },
    // reportHTML lays the results out with gobl.html's own stylesheets, as
    // if the benchmark were an invoice: one line per example, totals at the
    // end. The stylesheets are the ones the renderer links.
    async reportHTML(b) {
      if (!this.css) {
        const [env, inv] = await Promise.all(["/styles/envelope.css", "/styles/invoice.css"].map((u) => fetch(u).then((x) => x.text())));
        this.css = env + "\n" + inv;
      }
      const esc = (v) => String(v).replace(/[&<>"]/g, (c) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;" })[c]);
      const li = (label, value, cls = "") => `<li class="${cls}"><span class="label">${esc(label)}</span><span class="value">${esc(value)}</span></li>`;
      const lines = b.rows
        .map(
          (r, i) => `<tr>
            <td class="ref">${i + 1}</td>
            <td class="description"><div class="limited-height"><span>${esc(r.e.title.charAt(0).toUpperCase() + r.e.title.slice(1))}</span><br><small>${esc(r.e.file)}${r.e.addons?.length ? " · " + esc(r.e.addons.join(", ")) : ""}</small></div></td>
            <td class="quantity">${esc(r.kb)}</td>
            <td class="price">${esc(r.wasmMed)} ms</td>
            <td class="price">${esc(r.nativeMed)} ms</td>
            <td class="discount"><span class="discount-content">${esc(r.ratio)}</span></td>
            <td class="total">${esc(r.trip)} ms</td>
          </tr>`,
        )
        .join("");
      const t = b.total;
      return `<!DOCTYPE html>
<html data-theme="light"><head><meta charset="utf-8">
<link rel="preconnect" href="https://fonts.googleapis.com"><link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link href="https://fonts.googleapis.com/css2?family=Inter:wght@100..900&family=Geist+Mono:wght@100..900&display=swap" rel="stylesheet">
<style>${this.css}</style></head>
<body><article class="envelope"><article class="invoice">
  <header class="regular">
    <div class="details">
      <section class="title"><div class="title-group"><h1 class="type">Benchmark</h1><h2 class="code">WASM vs native</h2></div></section>
      <section class="summary"><h2 class="title">Summary</h2><ul>
        ${li("Date", b.date, "issue-date")}
        ${li("Examples", b.rows.length)}
        ${li("Rounds", b.rounds)}
        ${li("Unit", "milliseconds per render (median of rounds)")}
      </ul></section>
    </div>
    <div class="contacts">
      <section class="supplier"><div class="party-border"><h2>Browser · WASM</h2><div class="org-party">
        <div class="name">${esc(this.version ?? "gobl.html")}</div>
        <div>${esc(b.browser)}</div>
        <div>Render time measured around goblRender()</div>
      </div></div></section>
      <section class="customer"><div class="party-border"><h2>Server · native</h2><div class="org-party">
        <div class="name">gobl.html serve</div>
        <div>${esc(location.host)}</div>
        <div>Render time reported in the X-Render-Ms header</div>
      </div></div></section>
    </div>
  </header>
  <section class="lines"><h2>Lines</h2><table>
    <thead><tr><th class="ref">#</th><th class="description">Example</th><th class="quantity">KB</th><th class="price">WASM</th><th class="price">Native</th><th class="discount">Ratio</th><th class="total">Round trip</th></tr></thead>
    <tbody>${lines}</tbody>
  </table></section>
  <div class="tail">
    <div class="totals"><section class="totals"><h2>Totals</h2><table><tbody>
      <tr><th>WASM median</th><td>${esc(t.wasmMed)} ms</td></tr>
      <tr><th>WASM fastest</th><td>${esc(t.wasmMin)} ms</td></tr>
      <tr><th>Native median</th><td>${esc(t.nativeMed)} ms</td></tr>
      <tr><th>Native fastest</th><td>${esc(t.nativeMin)} ms</td></tr>
      <tr><th>Server round trip median</th><td>${esc(t.trip)} ms</td></tr>
      <tr class="sum strong"><th>Total WASM time</th><td>${esc(t.wasmSum)} ms</td></tr>
      <tr class="sum strong"><th>Total native time</th><td>${esc(t.nativeSum)} ms</td></tr>
      <tr class="payable strong"><th>WASM over native</th><td>${esc(t.ratio)}</td></tr>
    </tbody></table></section></div>
    <section class="notes"><h2>Notes</h2>
      <div class="note"><p>Both sides rendered the same documents with the same parameters, after one warm-up pass. Native times exclude request parsing and transfer; the round trip column is the full fetch from the browser.</p></div>
    </section>
  </div>
</article></article></body></html>`;
    },
  };
}
