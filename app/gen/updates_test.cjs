const assert = require("node:assert/strict");
const { readFileSync } = require("node:fs");
const { join } = require("node:path");
const { test } = require("node:test");
const vm = require("node:vm");

const origin = "https://webgo.test";
const resources = ["/", "/app.js", "/web/app.wasm",
  "https://external.test/rum.js"];
const workerSource = readFileSync(join(__dirname, "app-worker.js"), "utf8")
  .replaceAll("{{.Version}}", "new")
  .replaceAll("{{.ResourcesToCache}}", JSON.stringify(resources));
const updateSource = readFileSync(join(__dirname, "app-updates.js"), "utf8");
const tick = () => new Promise((resolve) => setImmediate(resolve));

function deferred() {
  let resolve, reject;
  const promise = new Promise((yes, no) => { resolve = yes; reject = no; });
  return { promise, resolve, reject };
}

function worker(overrides = {}) {
  const events = {};
  const calls = [];
  const cache = {
    addAll: async (requests) => { calls.push(["addAll", requests]); },
    match: async () => undefined,
    ...overrides.cache,
  };
  const caches = {
    open: async (name) => { calls.push(["open", name]); return cache; },
    keys: async () => ["app-old", "app-new", "unrelated"],
    delete: async (name) => { calls.push(["delete", name]); return true; },
    ...overrides.caches,
  };
  const self = {
    location: { origin },
    addEventListener: (name, callback) => { events[name] = callback; },
    skipWaiting: async () => { calls.push(["skipWaiting"]); },
    clients: { claim: async () => { calls.push(["claim"]); } },
  };
  const network = { source: "network" };
  vm.runInNewContext(workerSource, {
    self, caches, URL, Request,
    fetch: async (request) => { calls.push(["fetch", request.url]); return network; },
  });
  function lifecycle(name) {
    let promise;
    events[name]({ waitUntil: (value) => { promise = value; } });
    assert.equal(typeof promise?.then, "function", `${name} must use waitUntil synchronously`);
    return promise;
  }
  function fetchEvent(path, method = "GET") {
    let promise;
    events.fetch({
      request: new Request(new URL(path, origin), { method }),
      respondWith: (value) => { promise = value; },
    });
    return promise;
  }
  return { calls, lifecycle, fetchEvent, network };
}

test("install waits for precache before skipping the old worker", async () => {
  const pending = deferred();
  const w = worker({ cache: { addAll: () => pending.promise } });
  const installed = w.lifecycle("install");
  await tick();
  assert.equal(w.calls.some(([name]) => name === "skipWaiting"), false);
  pending.resolve();
  await installed;
  assert.equal(w.calls.at(-1)[0], "skipWaiting");
});

test("precache bypasses HTTP cache and excludes third-party scripts", async () => {
  const w = worker();
  await w.lifecycle("install");
  const requests = w.calls.find(([name]) => name === "addAll")[1];
  assert.equal(requests.length, 3);
  for (const request of requests) {
    assert.equal(new URL(request.url).origin, origin);
    assert.equal(request.cache, "reload");
    assert.equal(request.credentials, "same-origin");
    assert.equal(request.redirect, "error");
  }
});

test("failed install is rejected without activation or deleting the old cache", async () => {
  const w = worker({ cache: { addAll: async () => { throw Error("offline"); } } });
  await assert.rejects(w.lifecycle("install"), /offline/);
  assert.deepEqual(w.calls.filter(([name]) => name === "delete"), [["delete", "app-new"]]);
  assert.equal(w.calls.some(([name]) => name === "skipWaiting"), false);
});

test("activation waits for cleanup then claims clients; unrelated caches survive", async () => {
  const pending = deferred();
  const w = worker({ caches: {
    delete: async (name) => { w.calls.push(["delete", name]); await pending.promise; },
  } });
  const activated = w.lifecycle("activate");
  await tick();
  assert.equal(w.calls.some(([name]) => name === "claim"), false);
  assert.deepEqual(w.calls, [["delete", "app-old"]]);
  pending.resolve();
  await activated;
  assert.equal(w.calls.at(-1)[0], "claim");
});

test("fetch uses only this version's cache and retains offline app resources", async () => {
  const cached = { source: "new-cache" };
  const w = worker({ cache: { match: async () => cached } });
  assert.equal(await w.fetchEvent("/web/app.wasm"), cached);
  assert.deepEqual(w.calls, [["open", "app-new"]]);
});

test("cache miss fetches the network instead of searching stale caches", async () => {
  const w = worker();
  assert.equal(await w.fetchEvent("/app.js"), w.network);
  assert.deepEqual(w.calls, [["open", "app-new"], ["fetch", origin + "/app.js"]]);
});

test("runtime config, APIs, auth, navigations, third-party and writes bypass cache", () => {
  const w = worker();
  for (const path of ["/api/catalog", "/config/settings.json", "/config/version",
    "/auth/session", "/auth/login", "/devops", "/app-worker.js",
    "https://external.test/rum.js"]) {
    assert.equal(w.fetchEvent(path), undefined, path);
  }
  assert.equal(w.fetchEvent("/", "POST"), undefined);
  assert.deepEqual(w.calls, []);
});

function browser({ controlled = true, supported = true, update, auto = true } = {}) {
  const events = {};
  const state = { reloads: 0, updates: 0, logs: 0 };
  const workers = {
    controller: controlled ? {} : null,
    addEventListener: (name, callback) => { events[name] = callback; },
    getRegistration: async () => ({
      update: async () => { state.updates++; if (update) await update(); },
    }),
  };
  const navigator = { onLine: true };
  if (supported) navigator.serviceWorker = workers;
  const document = {
    visibilityState: "visible",
    addEventListener: (name, callback) => { events[name] = callback; },
  };
  const window = {
    location: { reload: () => { state.reloads++; } },
    addEventListener: (name, callback) => { events[name] = callback; },
    setInterval: (callback, delay) => { events.interval = callback; state.interval = delay; },
  };
  vm.runInNewContext(updateSource.replaceAll("{{.AutoUpdate}}", String(auto)), {
    navigator, document, window, console: { debug: () => { state.logs++; } },
  });
  return { events, state, workers, window, navigator, document };
}

test("replacement controller reloads once, including duplicate WASM notification", () => {
  const b = browser();
  b.workers.controller = {};
  b.events.controllerchange();
  b.events.controllerchange();
  b.window.goappReloadForUpdate();
  assert.equal(b.state.reloads, 1);
});

test("first installation does not reload; a subsequent update does", () => {
  const b = browser({ controlled: false });
  b.workers.controller = {};
  b.events.controllerchange();
  b.window.goappReloadForUpdate();
  assert.equal(b.state.reloads, 0);
  b.workers.controller = {};
  b.events.controllerchange();
  assert.equal(b.state.reloads, 1);
});

test("with auto update off a replacement controller does not reload", () => {
  const b = browser({ auto: false });
  b.workers.controller = {};
  b.events.controllerchange();
  b.window.goappReloadForUpdate();
  assert.equal(b.state.reloads, 0);
});

test("missing service-worker support is harmless", () => {
  const b = browser({ supported: false });
  assert.deepEqual(b.events, {});
  assert.equal(b.state.reloads, 0);
});

test("load, periodic, online and visibility events check for updates", async () => {
  const b = browser();
  for (const name of ["load", "interval", "online", "visibilitychange"]) {
    await b.events[name]();
  }
  assert.equal(b.state.updates, 4);
  assert.equal(b.state.interval, 60_000);
  assert.equal(b.state.reloads, 0);
});

test("offline and hidden tabs defer checks until online and visible again", async () => {
  const b = browser();
  b.navigator.onLine = false;
  await b.events.interval();
  b.navigator.onLine = true;
  b.document.visibilityState = "hidden";
  await b.events.online();
  assert.equal(b.state.updates, 0);
  b.document.visibilityState = "visible";
  await b.events.visibilitychange();
  assert.equal(b.state.updates, 1);
});

test("simultaneous checks share one in-flight update", async () => {
  const pending = deferred();
  const b = browser({ update: () => pending.promise });
  const checking = b.events.load();
  await tick();
  await b.events.online();
  await b.events.interval();
  assert.equal(b.state.updates, 1);
  pending.resolve();
  await checking;
  await b.events.interval();
  assert.equal(b.state.updates, 2);
});

test("failed update checks preserve the app and are retried", async () => {
  const b = browser({ update: async () => { throw Error("offline"); } });
  await b.events.load();
  await b.events.interval();
  assert.equal(b.state.updates, 2);
  assert.equal(b.state.logs, 2);
  assert.equal(b.state.reloads, 0);
});
