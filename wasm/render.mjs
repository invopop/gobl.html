// Renders GOBL envelopes through the WebAssembly build of gobl.html.
//
//   node render.mjs <gobl.html.wasm> <examples dir> <out dir> [params.json]
//
// Every *.json in the examples directory is rendered to <out dir>/<name>.html
// using the parameters that the Go tests derive from the file name, merged
// with the optional params file. Timing and failures go to stdout; the exit
// code is the number of failed renders.
//
// wasm_exec.js must sit next to the .wasm file. In Node the Go runtime needs
// a few browser globals, which are shimmed here.

import fs from "node:fs";
import path from "node:path";
import { performance } from "node:perf_hooks";
import { createRequire } from "node:module";

const [, , wasmPath, examplesDir, outDir, paramsPath] = process.argv;
if (!wasmPath || !examplesDir || !outDir) {
  console.error("usage: node render.mjs <gobl.html.wasm> <examples dir> <out dir> [params.json]");
  process.exit(2);
}

globalThis.require = createRequire(import.meta.url);
const events = new EventTarget();
globalThis.addEventListener = events.addEventListener.bind(events);
globalThis.dispatchEvent = events.dispatchEvent.bind(events);
await import(path.resolve(path.dirname(wasmPath), "wasm_exec.js"));

const go = new Go();
const started = performance.now();
const { instance } = await WebAssembly.instantiate(fs.readFileSync(wasmPath), go.importObject);
const ready = new Promise((resolve) => globalThis.addEventListener("gobl:ready", resolve));
go.run(instance);
await ready;
console.log(`instantiate+start: ${(performance.now() - started).toFixed(0)} ms`);
console.log(`version: ${JSON.stringify(globalThis.goblVersion())}`);

const base = paramsPath ? JSON.parse(fs.readFileSync(paramsPath, "utf8")) : {};

// paramsFor mirrors examples_test.go: the file name carries the state.
function paramsFor(name) {
  const p = { ...base };
  if (name.includes(".void.")) p.void = true;
  if (name.includes(".sandbox.")) p.sandbox = true;
  if (name.includes(".adjustment.")) p.adjustment_mode = true;
  for (const locale of ["es", "fr", "pl"]) {
    if (name.includes(`.${locale}.`)) p.locale = locale;
  }
  return p;
}

fs.mkdirSync(outDir, { recursive: true });
const files = fs.readdirSync(examplesDir).filter((f) => f.endsWith(".json")).sort();
const times = [];
let failed = 0;
for (const name of files) {
  const src = fs.readFileSync(path.join(examplesDir, name), "utf8");
  const t = performance.now();
  const res = globalThis.goblRender(src, paramsFor(name));
  times.push(performance.now() - t);
  if (res.error) {
    failed++;
    console.log(`FAIL ${name}: ${res.error}`);
    continue;
  }
  fs.writeFileSync(path.join(outDir, name.replace(/\.json$/, ".html")), res.html);
}
times.sort((a, b) => a - b);
const median = times[Math.floor(times.length / 2)] ?? 0;
console.log(
  `rendered ${files.length - failed}/${files.length}: median ${median.toFixed(1)} ms, ` +
    `max ${(times.at(-1) ?? 0).toFixed(1)} ms, total ${times.reduce((a, b) => a + b, 0).toFixed(0)} ms`,
);
process.exit(failed);
