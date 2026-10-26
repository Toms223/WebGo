const assert = require("node:assert/strict");
const { readFileSync } = require("node:fs");
const { join } = require("node:path");
const { test } = require("node:test");
const vm = require("node:vm");

const source = readFileSync(join(__dirname, "app.js"), "utf8")
  .replace("{{.Env}}", "{}")
  .replaceAll(/\{\{\.[A-Za-z]+\}\}/g, "")
  .replace("goappInitWebAssembly();\n", "");

const safari = "Mozilla/5.0 (iPhone; CPU iPhone OS 18_0 like Mac OS X) AppleWebKit/605.1.15 Version/18.0 Mobile/15E148 Safari/604.1";
const chrome = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/140.0 Safari/537.36";

function browser(userAgent) {
  const events = {};
  const context = {
    navigator: { userAgent, maxTouchPoints: 0 },
    window: {
      addEventListener: (name, fn) => { events[name] = fn; },
      matchMedia: () => ({ matches: false }),
    },
    document: { referrer: "" },
    console,
  };
  vm.runInNewContext(source, context);
  return { context, events };
}

test("an Apple browser is installable and showing the prompt does not throw", async () => {
  const b = browser(safari);
  assert.equal(b.context.goappIsAppInstallable(), true);
  await b.context.goappShowInstallPrompt();
});

test("a deferred install prompt is shown once", async () => {
  const b = browser(chrome);
  assert.equal(b.context.goappIsAppInstallable(), false);
  let prompted = 0;
  b.events.beforeinstallprompt({
    preventDefault() {},
    prompt() { prompted++; },
    userChoice: Promise.resolve({ outcome: "accepted" }),
  });
  assert.equal(b.context.goappIsAppInstallable(), true);
  await b.context.goappShowInstallPrompt();
  assert.equal(prompted, 1);
  assert.equal(b.context.goappIsAppInstallable(), false);
  await b.context.goappShowInstallPrompt();
  assert.equal(prompted, 1);
});
