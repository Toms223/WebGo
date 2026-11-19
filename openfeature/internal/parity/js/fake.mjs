import { OpenFeatureEventEmitter, TypedInMemoryProvider, createEvaluationEvent, instantiateErrorByErrorCode } from '@openfeature/web-sdk';

export function makeError(e) {
  const err = e.code ? instantiateErrorByErrorCode(e.code, e.message) : new Error(e.message);
  err.stack = String(err);
  return err;
}

export function errorRecord(err) {
  return { code: err?.code, message: err?.message };
}

export function toContext(v) {
  if (Array.isArray(v)) return v.map(toContext);
  if (v && typeof v === 'object') {
    if (Object.keys(v).length === 1 && typeof v.$date === 'string') return new Date(v.$date);
    return Object.fromEntries(Object.entries(v).map(([k, x]) => [k, toContext(x)]));
  }
  return v;
}

function evaluator(def, defaultVariant) {
  if (def.match) {
    return (ctx) => (Object.entries(def.match).every(([k, v]) => ctx[k] === v) ? def.variant : defaultVariant);
  }
  return (ctx) => {
    const v = ctx[def.attr];
    return typeof v === 'string' && Object.hasOwn(def.map, v) ? def.map[v] : defaultVariant;
  };
}

export function memoryFlags(flags) {
  const out = {};
  for (const [key, f] of Object.entries(flags)) {
    const flag = { variants: f.variants, defaultVariant: f.defaultVariant };
    if (f.disabled !== undefined) flag.disabled = f.disabled;
    if (f.contextEvaluator) flag.contextEvaluator = evaluator(f.contextEvaluator, f.defaultVariant);
    out[key] = flag;
  }
  return out;
}

function trimArgs(args) {
  const a = [...args];
  while (a.length && a[a.length - 1] === undefined) a.pop();
  return a;
}

const types = { boolean: 'Boolean', string: 'String', number: 'Number', object: 'Object' };

export function makeProvider(id, def, env) {
  if (def.memory) return new TypedInMemoryProvider(memoryFlags(def.memory));
  const counts = {};
  const call = (method, args) => {
    counts[method] = (counts[method] || 0) + 1;
    env.record({ kind: 'provider', p: id, call: method, args: trimArgs(args) });
    return `${id}.${method}#${counts[method]}`;
  };
  const settle = (mode, key) => (mode === 'deferred' ? env.defer(key) : Promise.resolve());
  const p = { metadata: { ...(def.metadata || {}), name: def.name ?? id } };
  if (def.runsOn !== undefined) p.runsOn = def.runsOn;
  if (def.domainScoped !== undefined) p.domainScoped = def.domainScoped;
  if (def.events) p.events = new OpenFeatureEventEmitter();
  if (def.hooks) p.hooks = def.hooks.map((h) => env.hook(h));
  if (def.initialize && def.initialize !== 'none') {
    p.initialize = (ctx, domain) => settle(def.initialize, call('initialize', [ctx, domain]));
  }
  if (def.onContextChange && def.onContextChange !== 'none') {
    p.onContextChange = (oldCtx, newCtx) => {
      const key = call('onContextChange', [oldCtx, newCtx]);
      if (def.onContextChange === 'throw') throw makeError({ message: `${id} onContextChange throw` });
      if (def.onContextChange === 'deferred') return env.defer(key);
    };
  }
  if (def.onClose && def.onClose !== 'none') {
    p.onClose = () => settle(def.onClose, call('onClose', []));
  }
  if (def.track) {
    p.track = (name, ctx, details) => {
      call('track', [name, ctx, details]);
      if (def.track === 'throw') throw makeError({ message: `${id} track throw` });
    };
  }
  for (const [type, T] of Object.entries(types)) {
    const method = `resolve${T}Evaluation`;
    p[method] = (flagKey, defaultValue, ctx) => {
      call(method, [flagKey, defaultValue, ctx]);
      const spec = def.flags?.[flagKey]?.[type];
      if (!spec) return { value: defaultValue, reason: 'DEFAULT' };
      if (spec.throw) throw makeError(spec.throw);
      return { value: defaultValue, ...spec };
    };
  }
  return p;
}

export function makeHook(id, def, env) {
  const stage = (name, hc, hints, extra) => {
    const r = {
      kind: 'hook', h: id, stage: name,
      flagKey: hc.flagKey, defaultValue: hc.defaultValue, flagValueType: hc.flagValueType,
      context: hc.context, domain: hc.clientMetadata.domain, providerName: hc.providerMetadata.name,
      hints, ...extra,
    };
    if (def.data) {
      r.data = hc.hookData.get('n') ?? 0;
      hc.hookData.set('n', r.data + 1);
    }
    env.record(r);
    if (def.telemetry && name === 'finally') env.record({ kind: 'telemetry', event: createEvaluationEvent(hc, extra.details) });
    if (def.throwIn?.includes(name)) throw makeError({ message: `hook ${id} ${name}` });
  };
  return {
    before: (hc, hints) => stage('before', hc, hints),
    after: (hc, details, hints) => stage('after', hc, hints, { details }),
    error: (hc, err, hints) => stage('error', hc, hints, { error: errorRecord(err) }),
    finally: (hc, details, hints) => stage('finally', hc, hints, { details }),
  };
}
