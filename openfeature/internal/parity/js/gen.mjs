import { readdirSync, readFileSync, unlinkSync, writeFileSync } from 'node:fs';
import { OpenFeatureAPI } from '@openfeature/web-sdk';
import { errorRecord, makeError, makeHook, makeProvider, memoryFlags, toContext } from './fake.mjs';

const root = new URL('..', import.meta.url).pathname;

export function canon(v) {
  if (v === null) return 'null';
  if (v instanceof Date) return JSON.stringify(v.toISOString());
  if (Array.isArray(v)) return `[${v.map((x) => (x === undefined || typeof x === 'function' ? 'null' : canon(x))).join(',')}]`;
  if (typeof v === 'object') {
    return `{${Object.keys(v).sort()
      .filter((k) => v[k] !== undefined && typeof v[k] !== 'function')
      .map((k) => `${JSON.stringify(k)}:${canon(v[k])}`).join(',')}}`;
  }
  return JSON.stringify(v);
}

const tick = () => new Promise((r) => setTimeout(r, 0));
const dom = (d, s = {}) => (d === '' && !s.literalDomain ? undefined : d);
const ofTarget = (t, s) => (t === 'api' ? null : dom(t.slice('client:'.length), s));

async function run(sc) {
  const trace = [];
  const record = (r) => trace.push(canon(r));
  const deferred = new Map();
  const hooks = {};
  const env = {
    record,
    defer: (key) => new Promise((resolve, reject) => deferred.set(key, { resolve, reject })),
    hook: (id) => hooks[id],
  };
  for (const [id, def] of Object.entries(sc.hooks || {})) hooks[id] = makeHook(id, def, env);
  const providers = {};
  for (const [id, def] of Object.entries(sc.providers || {})) providers[id] = makeProvider(id, def, env);

  const api = new OpenFeatureAPI();
  const str = (a) => (a instanceof Error ? a.message : String(a));
  const logLine = (level) => (...args) => {
    const r = { kind: 'log', level, msg: str(args[0]) };
    if (args.length > 1) r.args = args.slice(1).map(str);
    record(r);
  };
  api.setLogger({ error: logLine('error'), warn: logLine('warn'), info: logLine('info'), debug: logLine('debug') });

  const clients = new Map();
  const client = (d) => {
    const key = d === undefined ? '\0' : d;
    if (!clients.has(key)) clients.set(key, api.getClient(d));
    return clients.get(key);
  };
  const handlers = {};
  const ret = (op, fn) => {
    let p;
    try {
      p = fn();
    } catch (err) {
      record({ kind: 'return', op, error: errorRecord(err) });
      return;
    }
    if (p instanceof Promise) {
      p.then(() => record({ kind: 'return', op }), (err) => record({ kind: 'return', op, error: errorRecord(err) }));
    } else {
      record({ kind: 'return', op });
    }
  };
  const bind = (s, fnName) => {
    const provider = providers[s.provider];
    if (!provider) throw new Error(`unknown provider ${s.provider}`);
    const ctx = s.context === undefined ? [] : [toContext(s.context)];
    const d = dom(s.domain, s);
    return d === undefined ? api[fnName](provider, ...ctx) : api[fnName](d, provider, ...ctx);
  };
  const withDomain = (d, rest) => (d === undefined ? rest : [d, ...rest]);

  const ops = {
    setProvider: (s) => ret('setProvider', () => { bind(s, 'setProvider'); }),
    setProviderAndWait: (s) => ret('setProviderAndWait', () => bind(s, 'setProviderAndWait')),
    resolve: (s) => settle(s, 'resolve'),
    reject: (s) => settle(s, 'reject'),
    setContext: (s) => ret('setContext', () => api.setContext(...withDomain(dom(s.domain), [toContext(s.context)]))),
    clearContext: (s) => ret('clearContext', () => api.clearContext(dom(s.domain))),
    clearContexts: () => ret('clearContexts', () => api.clearContexts()),
    getContext: (s) => record({ kind: 'context', domain: dom(s.domain), context: api.getContext(dom(s.domain)) }),
    addHooks: (s) => level(s.level).addHooks(...s.hooks.map((h) => hooks[h])),
    clearHooks: (s) => level(s.level).clearHooks(),
    addHandler: (s) => {
      const target = s.target === 'api' ? api : s.fresh ? api.getClient(ofTarget(s.target, s)) : client(ofTarget(s.target, s));
      const fn = (details) => {
        record({ kind: 'event', target: s.target, handler: s.handler, event: s.event, details: { ...details } });
        if (s.readStatus !== undefined) {
          record({ kind: 'status', domain: dom(s.readStatus), status: client(dom(s.readStatus)).providerStatus });
        }
        if (s.throws) throw makeError({ message: `handler ${s.handler}` });
      };
      handlers[s.handler] = { target, event: s.event, fn };
      target.addHandler(s.event, fn);
    },
    removeHandler: (s) => {
      const h = handlers[s.handler];
      h.target.removeHandler(h.event, h.fn);
    },
    emit: (s) => providers[s.provider].events.emit(s.event, s.details),
    evaluate: (s) => {
      const T = s.type[0].toUpperCase() + s.type.slice(1);
      const options = {};
      if (s.hooks) options.hooks = s.hooks.map((h) => hooks[h]);
      if (s.hints) options.hookHints = s.hints;
      const details = client(dom(s.domain))[`get${T}Details`](s.flag, s.default, options);
      record({ kind: 'result', details });
    },
    track: (s) => client(dom(s.domain)).track(s.name, ...(s.details === undefined ? [] : [s.details])),
    status: (s) => record({ kind: 'status', domain: dom(s.domain, s), status: client(dom(s.domain, s)).providerStatus }),
    clientMetadata: (s) => {
      const c = s.version === undefined ? client(dom(s.domain, s)) : api.getClient(dom(s.domain, s), s.version);
      const { sdk, ...value } = c.metadata;
      record({ kind: 'metadata', which: 'client', value });
    },
    providerMetadata: (s) => record({ kind: 'metadata', which: 'provider', value: api.getProviderMetadata(dom(s.domain)) }),
    putConfiguration: (s) => { providers[s.provider].putConfiguration(memoryFlags(s.flags)); },
    close: () => ret('close', () => api.close()),
    clearProviders: () => ret('clearProviders', () => api.clearProviders()),
  };
  function level(l) {
    return l === 'api' ? api : client(ofTarget(l));
  }
  function settle(s, how) {
    const d = deferred.get(s.target);
    if (!d) throw new Error(`no pending call ${s.target}`);
    deferred.delete(s.target);
    if (how === 'resolve') d.resolve();
    else d.reject(makeError(s.error));
  }

  for (const [i, s] of sc.steps.entries()) {
    if (!ops[s.op]) throw new Error(`unknown op ${s.op}`);
    record({ kind: 'step', i, op: s.op });
    ops[s.op](s);
    await tick();
    await tick();
  }
  return trace;
}

const files = readdirSync(`${root}scenarios`).filter((f) => f.endsWith('.json')).sort();
let failed = false;
for (const g of readdirSync(`${root}golden`)) {
  if (g.endsWith('.trace.json') && !files.includes(g.replace(/\.trace\.json$/, '.json'))) unlinkSync(`${root}golden/${g}`);
}
for (const f of files) {
  const sc = JSON.parse(readFileSync(`${root}scenarios/${f}`, 'utf8'));
  if (`${sc.id}.json` !== f) throw new Error(`${f}: id ${sc.id} does not match file name`);
  try {
    const trace = await run(sc);
    if (!trace.some((r) => /^\{.*"kind":"(result|event|status)"/.test(r))) throw new Error('no result, event or status record');
    writeFileSync(`${root}golden/${sc.id}.trace.json`, `[\n${trace.join(',\n')}\n]\n`);
  } catch (err) {
    failed = true;
    console.error(`${f}: ${err.stack}`);
  }
}
console.log(`${files.length} scenarios`);
process.exit(failed ? 1 : 0);
