import http from 'k6/http';
import { check } from 'k6';

export const options = {
  scenarios: {
    spike: {
      executor: 'ramping-arrival-rate',
      timeUnit: '1s',
      startRate: 800,
      preAllocatedVUs: 400,
      maxVUs: 4000,
      stages: [
        { target: 1500, duration: '30s' }, // 800 -> 1500
        { target: 7000, duration: '5s'  }, // jump
        { target: 7000, duration: '1m'  }, // hold
        { target: 800,  duration: '1m'  }, // cool down
        { target: 0,    duration: '10s' }, // end
      ],
    },
  },
  thresholds: {
    http_req_failed: ['rate<0.05'],
    http_req_duration: ['p(95)<2000'],
  },
};

const BASE = __ENV.TARGET_BASE_URL || 'http://load-target.load-test.svc.cluster.local';

export default function () {
  const res = http.post(`${BASE}/work?cpu_ms=10&mem_mb=5`, null, { timeout: '10s' });
  check(res, { 'status is 200': (r) => r.status === 200 });
}
