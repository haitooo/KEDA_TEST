import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = { vus: 1, duration: '15s' };

export default function () {
  const res = http.get('http://34.146.83.214/');
  check(res, { 'status is 200': (r) => r.status === 200 });
  sleep(1);
}
