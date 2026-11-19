import { readFileSync } from 'node:fs';
import * as sdk from '@openfeature/web-sdk';

const surface = JSON.parse(readFileSync(new URL('../surface.json', import.meta.url), 'utf8'));
const errors = [];
const open = [];
const has = (obj, name) => obj != null && (name in Object(obj) || (typeof obj === 'function' && name in obj.prototype));
const get = (obj, name) => (name in Object(obj) ? obj[name] : obj.prototype[name]);

for (const [key, entry] of Object.entries(surface)) {
  const kinds = ['go', 'deviation', 'absent', 'internal', 'open'].filter((k) => k in entry);
  if (kinds.length !== 1) errors.push(`${key}: needs exactly one of go, deviation, absent, internal, open`);
  if (entry.deviation && !/^D[1-8]$/.test(entry.deviation)) errors.push(`${key}: unknown deviation ${entry.deviation}`);
  if (entry.open) open.push(`${key}: ${entry.open}`);
  if (key.startsWith('type:')) continue;
  const [head, ...rest] = key.split('.');
  let obj = head === 'Client' ? sdk.OpenFeature.getClient() : sdk[head];
  if (obj === undefined) {
    errors.push(`${key}: no export ${head}`);
    continue;
  }
  for (const seg of rest) {
    if (!has(obj, seg)) {
      errors.push(`${key}: missing member ${seg}`);
      obj = undefined;
      break;
    }
    obj = get(obj, seg);
  }
  if (obj === undefined) continue;
  if ('value' in entry && obj !== entry.value) errors.push(`${key}: value ${obj} != ${entry.value}`);
  if ('code' in entry && new obj('x').code !== entry.code) errors.push(`${key}: code ${new obj('x').code} != ${entry.code}`);
}
for (const name of Object.keys(sdk)) {
  if (!(name in surface)) errors.push(`export ${name} is not listed`);
}
const instances = Object.entries(sdk).filter(([, v]) => v !== null && typeof v === 'object');
const listed = (cls, m) => {
  const e = surface[cls] || {};
  return e.deviation || e.internal || `${cls}.${m}` in surface
    || instances.some(([n, v]) => v instanceof sdk[cls] && `${n}.${m}` in surface);
};
let members = 0;
for (const [name, cls] of Object.entries(sdk)) {
  if (typeof cls !== 'function' || !cls.prototype) continue;
  for (let p = cls.prototype; p && p !== Object.prototype && p !== Error.prototype; p = Object.getPrototypeOf(p)) {
    for (const m of Object.getOwnPropertyNames(p)) {
      if (m === 'constructor' || m.startsWith('_')) continue;
      members++;
      if (!listed(name, m)) errors.push(`member ${name}.${m} is not listed`);
    }
  }
}
for (const o of open) errors.push(`open entry ${o}`);
for (const e of errors) console.error(`FAIL ${e}`);
console.log(`${Object.keys(surface).length} entries, ${members} class members, ${open.length} open, ${errors.length} failures`);
process.exit(errors.length ? 1 : 0);
