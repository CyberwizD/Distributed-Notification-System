import http from 'k6/http';
import { check } from 'k6';
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const AUTH_TOKEN = __ENV.API_TOKEN || 'test-token';
const TEST_USER_ID = __ENV.TEST_USER_ID || '00000000-0000-0000-0000-000000000001';

export const options = {
  scenarios: {
    notifications: {
      executor: 'constant-arrival-rate',
      duration: __ENV.TEST_DURATION || '2m',
      rate: Number(__ENV.REQUEST_RATE) || 1000,
      timeUnit: '1m',
      preAllocatedVUs: Number(__ENV.PRE_ALLOCATED_VUS) || 50,
      maxVUs: Number(__ENV.MAX_VUS) || 200,
    },
  },
  thresholds: {
    http_req_duration: ['p(95)<1000'],
    http_req_failed: ['rate<0.01'],
  },
};

export default function () {
  const payload = {
    request_id: uuidv4(),
    user_id: TEST_USER_ID,
    channel: 'push',
    template_slug: __ENV.TEMPLATE_SLUG || 'welcome_notification',
    variables: {
      name: 'Load Tester',
      link: 'https://example.com/reset',
    },
    priority: 'normal',
    metadata: {
      correlation: uuidv4(),
    },
  };

  const res = http.post(`${BASE_URL}/v1/notifications/send`, JSON.stringify(payload), {
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${AUTH_TOKEN}`,
      'X-Correlation-ID': payload.metadata.correlation,
    },
  });

  check(res, {
    'queued or ok': (r) => r.status === 202 || r.status === 200,
  });
}
