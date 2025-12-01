import http from 'k6/http';
import { check, sleep } from 'k6';
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

// Configuration
export const options = {
  stages: [
    { duration: '10s', target: 20 }, // Ramp up to 20 users
    { duration: '30s', target: 20 }, // Stay at 20 users
    { duration: '10s', target: 0 },  // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'], // 95% of requests should be below 500ms
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:4000';

// Setup: Create two accounts to transfer between
export function setup() {
  const headers = { 'Content-Type': 'application/json' };
  
  const accountRes1 = http.post(`${BASE_URL}/accounts`, JSON.stringify({ currency: 'USD' }), { headers });
  const accountRes2 = http.post(`${BASE_URL}/accounts`, JSON.stringify({ currency: 'USD' }), { headers });

  const account1 = accountRes1.json('account_id');
  const account2 = accountRes2.json('account_id');

  if (!account1 || !account2) {
    console.log('Account 1 Response:', accountRes1.body);
    console.log('Account 2 Response:', accountRes2.body);
    throw new Error('Failed to create accounts for testing');
  }

  return { account1, account2 };
}

export default function (data) {
  const { account1, account2 } = data;
  const headers = { 'Content-Type': 'application/json' };
  
  const payload = JSON.stringify({
    from_account_id: account1,
    to_account_id: account2,
    amount_minor: 10, // 10 cents
    currency: 'USD',
    idempotency_key: uuidv4(), // Unique key for every request
  });

  const res = http.post(`${BASE_URL}/transfers`, payload, { headers });

  check(res, {
    'is status 202': (r) => r.status === 202,
  });

  sleep(0.01); // Small sleep to limit rate per VU
}
