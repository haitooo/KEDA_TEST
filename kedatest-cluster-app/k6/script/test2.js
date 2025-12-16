import http from 'k6/http';
import { check } from 'k6';

export const options = {
  scenarios: {
    slow_ramp: {
      executor: 'ramping-arrival-rate',
      timeUnit: '1s',
      startRate: 100,
      preAllocatedVUs: 200,
      maxVUs: 2000,
      stages: [
        { target: 1000, duration: '2m' },
        { target: 1000, duration: '30s' },
        { target: 0,    duration: '20s' },
      ],
    },
  },
};

const BASE = __ENV.TARGET_BASE_URL || 'http://load-target.load-test.svc.cluster.local';

export default function () {
  const res = http.post(`${BASE}/work?cpu_ms=10&mem_mb=5`, null, { timeout: '10s' });
  check(res, { 'status is 200': (r) => r.status === 200 });
}
