const assert = require("node:assert/strict");
const { spawn } = require("node:child_process");
const { readFileSync, mkdtempSync, rmSync } = require("node:fs");
const http = require("node:http");
const { tmpdir } = require("node:os");
const { join } = require("node:path");
const { test } = require("node:test");

const lifecycle = readFileSync(join(__dirname, "app-worker.js"), "utf8");
const bootstrap = readFileSync(join(__dirname, "app-updates.js"), "utf8")
  .replaceAll("{{.AutoUpdate}}", "true");
const resources = ["/", "/app.js", "/web/app.wasm"];
const delay = (ms) => new Promise((resolve) => setTimeout(resolve, ms));

async function until(check, label) {
  const deadline = Date.now() + 15_000;
  while (Date.now() < deadline) {
    if (await check()) return;
    await delay(100);
  }
  throw Error("Timed out: " + label);
}

test("Chrome migrates a legacy cache, automatically upgrades, and retries failed installs", {
  skip: !process.env.CHROME_BIN || typeof WebSocket === "undefined",
  timeout: 60_000,
}, async () => {
  let version = "legacy";
  let failWasm = false;
  let failedFetches = 0;
  const server = http.createServer((request, response) => {
    response.setHeader("Cache-Control", "no-cache");
    if (request.url === "/app-worker.js") {
      response.setHeader("Content-Type", "application/javascript");
      if (version === "legacy") {
        response.end(`
          const name = "app-legacy";
          self.addEventListener("install", e => e.waitUntil((async () => {
            const cache = await caches.open(name);
            await cache.addAll(${JSON.stringify(resources)});
            await self.skipWaiting();
          })()));
          self.addEventListener("activate", e => e.waitUntil(self.clients.claim()));
          self.addEventListener("fetch", e => e.respondWith((async () =>
            await caches.match(e.request) || fetch(e.request))()));
        `);
      } else {
        response.end(lifecycle.replaceAll("{{.Version}}", version)
          .replaceAll("{{.ResourcesToCache}}", JSON.stringify(resources)));
      }
    } else if (request.url === "/app.js") {
      response.setHeader("Content-Type", "application/javascript");
      response.end(`document.body.dataset.jsVersion = ${JSON.stringify(version)};`);
    } else if (request.url === "/web/app.wasm") {
      if (failWasm) {
        failedFetches++;
        response.writeHead(503);
      }
      response.end(version);
    } else {
      response.setHeader("Content-Type", "text/html");
      response.end(`<!doctype html><html><head>
        <script>${version === "legacy" ? "" : bootstrap}</script>
        <script defer src="/app.js"></script>
        </head><body><script>
        sessionStorage.loads = String(Number(sessionStorage.loads || 0) + 1);
        fetch("/web/app.wasm").then(r => r.text()).then(version => {
          document.body.dataset.version = version;
        });
        navigator.serviceWorker.register("/app-worker.js").then(reg => {
          ${version !== "legacy" ? "" : `
          reg.addEventListener("updatefound", () => {
            const worker = reg.installing;
            worker.addEventListener("statechange", () => {
              if (worker.state === "activated" && navigator.serviceWorker.controller) {
                location.reload();
              }
            });
          });`}
        });
        </script></body></html>`);
    }
  });
  await new Promise((resolve) => server.listen(0, "localhost", resolve));
  const origin = `http://localhost:${server.address().port}`;
  const profile = mkdtempSync(join(tmpdir(), "webgo-worker-chrome-"));
  const chrome = spawn(process.env.CHROME_BIN, [
    "--headless", "--disable-gpu", "--no-first-run", "--no-default-browser-check",
    "--remote-debugging-port=0", `--user-data-dir=${profile}`, "about:blank",
  ], { stdio: ["ignore", "ignore", "pipe"] });
  let socket;
  try {
    let debugURL;
    let stderr = "";
    chrome.stderr.on("data", (chunk) => {
      stderr += chunk;
      debugURL = stderr.match(/DevTools listening on (ws:\/\/[^\s]+)/)?.[1];
    });
    await until(() => debugURL, "Chrome debugger startup");
    const targets = await (await fetch(
      `http://${new URL(debugURL).host}/json/list`
    )).json();
    socket = new WebSocket(targets.find((target) => target.type === "page").webSocketDebuggerUrl);
    await new Promise((resolve, reject) => {
      socket.addEventListener("open", resolve, { once: true });
      socket.addEventListener("error", reject, { once: true });
    });
    let nextID = 0;
    const pending = new Map();
    socket.addEventListener("message", (event) => {
      const result = JSON.parse(event.data);
      const waiter = pending.get(result.id);
      if (waiter) {
        pending.delete(result.id);
        result.error ? waiter.reject(Error(result.error.message)) : waiter.resolve(result.result);
      }
    });
    function command(method, params = {}) {
      return new Promise((resolve, reject) => {
        const id = ++nextID;
        pending.set(id, { resolve, reject });
        socket.send(JSON.stringify({ id, method, params }));
      });
    }
    async function evaluate(expression) {
      const result = await command("Runtime.evaluate", {
        expression, awaitPromise: true, returnByValue: true,
      });
      if (result.exceptionDetails) throw Error(JSON.stringify(result.exceptionDetails));
      return result.result.value;
    }
    async function loaded(want) {
      try {
        return await evaluate(`document.body?.dataset.version === ${JSON.stringify(want)} &&
          document.body?.dataset.jsVersion === ${JSON.stringify(want)} &&
          !!navigator.serviceWorker.controller`);
      } catch {
        return false;
      }
    }
    await command("Page.navigate", { url: origin + "/" });
    await until(() => loaded("legacy"), "legacy version installed");
    const legacyLoads = await evaluate("Number(sessionStorage.loads)");
    version = "v2";
    await evaluate("navigator.serviceWorker.getRegistration().then(r => r.update())");
    await until(() => loaded("v2"), "legacy tab reloads onto v2");
    assert.equal(await evaluate("Number(sessionStorage.loads)"), legacyLoads + 1);
    assert.deepEqual(await evaluate("caches.keys()"), ["app-v2"]);

    version = "v3";
    failWasm = true;
    await evaluate('window.dispatchEvent(new Event("online"))');
    await until(() => failedFetches > 0, "failed update attempted automatically");
    await until(async () => !(await evaluate(
      "navigator.serviceWorker.getRegistration().then(r => !!r.installing)"
    )), "failed install finishes");
    assert.ok(await loaded("v2"), "failed update leaves v2 running");
    assert.deepEqual(await evaluate("caches.keys()"), ["app-v2"]);

    failWasm = false;
    await evaluate('document.dispatchEvent(new Event("visibilitychange"))');
    await until(() => loaded("v3"), "visible tab automatically retries and reloads v3");
    assert.deepEqual(await evaluate("caches.keys()"), ["app-v3"]);
    assert.equal(await evaluate("Number(sessionStorage.loads)"), legacyLoads + 2);

    await command("Storage.clearDataForOrigin", {
      origin, storageTypes: "service_workers,cache_storage",
    });
    const loads = await evaluate("Number(sessionStorage.loads)");
    await command("Page.navigate", { url: origin + "/" });
    await until(() => loaded("v3"), "fresh install claims page");
    await delay(500);
    assert.equal(await evaluate("Number(sessionStorage.loads)"), loads + 1);
  } finally {
    socket?.close();
    const stopped = new Promise((resolve) => chrome.once("exit", resolve));
    chrome.kill();
    await stopped;
    await new Promise((resolve) => server.close(resolve));
    rmSync(profile, { recursive: true, force: true });
  }
});
