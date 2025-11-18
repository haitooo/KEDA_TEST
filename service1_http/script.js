import http from 'k6/http';
import { check } from 'k6';

export const options = {
  scenarios: {
    spike_test: {
      executor: 'ramping-arrival-rate',
      startRate: 1,        // 最初のリクエストレート
      timeUnit: '1s',      // rate の単位
      preAllocatedVUs: 100,
      maxVUs: 400,
      stages: [
        { target: 100, duration: '5s' }, // 1  -> 100
        { target: 500, duration: '10s' }, // 100 -> 500
        { target: 1000, duration: '5s' }, // 500 -> 1000
        { target: 1000, duration: '180s' }, // 1000 を 60秒維持
        { target: 100, duration: '10s' }, // 1000 -> 100
        { target: 1,    duration: '5s' }, // 1000 -> 1
      ],
    },
  },
  thresholds: {
    http_req_failed: ['rate<0.05'],
    http_req_duration: ['p(95)<300'],
  },
};

const params = {
  timeout: '500ms',
};

export default function () {
  const res = http.get('http://<IP address>', params);
  check(res, { 'status is 200': (r) => r.status === 200 });
}
