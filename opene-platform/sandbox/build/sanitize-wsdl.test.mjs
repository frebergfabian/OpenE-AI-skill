import {test} from 'node:test';
import assert from 'node:assert/strict';
import {sanitizeWsdl} from './sanitize-wsdl.mjs';

test('replaces remote endpoints and imports without changing namespaces or local imports', () => {
  const input = `<definitions xmlns:s="https://schemas.example.org/type"><s:address location="https://vendor.invalid/service"/><s:import location='https://vendor.invalid/contract'/><s:import location="local.xsd"/></definitions>`;
  const {result, changed} = sanitizeWsdl(input);
  assert.equal(changed, 2);
  assert.ok(!result.includes('vendor.invalid'));
  assert.match(result, /xmlns:s="https:\/\/schemas\.example\.org\/type"/);
  assert.match(result, /location="local\.xsd"/);
  assert.equal((result.match(/http:\/\/127\.0\.0\.1:1\/disabled/g) || []).length, 2);
  assert.deepEqual(sanitizeWsdl(result), {result, changed: 0});
});

test('rejects contracts encoded with embedded NUL bytes', () => {
  assert.throws(() => sanitizeWsdl('<\0d\0:address>'), /encoding/);
});
