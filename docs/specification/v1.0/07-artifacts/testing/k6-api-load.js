// ECD-013 load test — API read/write mix at RFC-014 G.3 targets. Budgets asserted, not observed.
import http from 'k6/http';
import { check } from 'k6';
export const options = {
  scenarios: {
    reads:  { executor: 'constant-arrival-rate', rate: 300, timeUnit: '1s', duration: '10m', preAllocatedVUs: 200 },
    gates:  { executor: 'constant-arrival-rate', rate: 5,   timeUnit: '1s', duration: '10m', preAllocatedVUs: 50 },
  },
  thresholds: {
    'http_req_duration{kind:list}':   ['p(95)<250'],   // list endpoints p95 250ms budget
    'http_req_duration{kind:detail}': ['p(95)<120'],
    'http_req_duration{kind:gate}':   ['p(95)<25000'], // CI gate sync budget
    http_req_failed: ['rate<0.001'],                   // 99.9% availability
  },
};
const H = { Authorization: `Bearer ${__ENV.TOKEN}` };
export default function () {
  const base = __ENV.BASE || 'https://staging-api.nydux.ai/v1';
  let r = http.get(`${base}/kernels?limit=100&sort=-kes`, { headers: H, tags: { kind: 'list' } });
  check(r, { 'list 200': (x) => x.status === 200 });
  const items = r.json('items');
  if (items && items.length) {
    r = http.get(`${base}/kernels/${items[0].kernel_hash}`, { headers: H, tags: { kind: 'detail' } });
    check(r, { 'detail 200': (x) => x.status === 200 });
  }
  if (__ITER % 60 === 0) {
    r = http.post(`${base}/regressions/checks`, JSON.stringify({
      from: { fingerprint: __ENV.FP_FROM }, to: { fingerprint: __ENV.FP_TO }, fail_on: 'CRI>0.10',
    }), { headers: { ...H, 'Content-Type': 'application/json' }, tags: { kind: 'gate' } });
    check(r, { 'gate 200/202': (x) => x.status === 200 || x.status === 202 });
  }
}
