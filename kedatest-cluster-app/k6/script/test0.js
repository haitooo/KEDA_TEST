import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  scenarios: {
    smoke: {
      executor: 'ramping-arrival-rate',
      timeUnit: '1s',
      startRate: 5,
      preAllocatedVUs: 20,
      maxVUs: 200,
      stages: [
        { target: 20, duration: '30s' },
        { target: 50, duration: '30s' },
        { target: 0,  duration: '10s' },
      ],
    },
  },
};

const BASE = __ENV.TARGET_BASE_URL || 'http://load-target.load-test.svc.cluster.local';

export default function () {
  const res = http.post(`${BASE}/work?cpu_ms=5&mem_mb=1`, null, { timeout: '10s' });
  check(res, { 'status is 200': (r) => r.status === 200 });
  sleep(0.01);
}
