// k6 load test: GET /api/rooms/{roomId}/slots/list (hottest endpoint per TASK.md).
// Prerequisites: API up (e.g. docker compose), DB seeded (`make seed`) so room + schedule exist.
//
// Env:
//   BASE_URL           — default http://127.0.0.1:8080
//   LOADTEST_ROOM_ID   — default seed room "Conference Alpha"
//   LOADTEST_RATE      — requests per second (default 50; try 100 per task)
//   LOADTEST_DURATION  — e.g. 90s, 2m (default 90s)
//   LOADTEST_PROFILE=smoke — short 5 VUs × 10s (sanity check)

import http from "k6/http";
import { check, sleep } from "k6";

const BASE_URL = (__ENV.BASE_URL || "http://127.0.0.1:8080").replace(/\/$/, "");
const ROOM_ID =
  __ENV.LOADTEST_ROOM_ID || "11111111-1111-4111-8111-111111111101";

function nextWeekdayUTCISO() {
  const now = Date.now();
  for (let i = 1; i <= 14; i++) {
    const x = new Date(now + i * 86400000);
    const wd = x.getUTCDay();
    const apiWd = wd === 0 ? 7 : wd;
    if (apiWd >= 1 && apiWd <= 5) {
      return x.toISOString().slice(0, 10);
    }
  }
  return "2030-06-03";
}

const smoke = __ENV.LOADTEST_PROFILE === "smoke";
const rate = Number(__ENV.LOADTEST_RATE || 50);
const duration = __ENV.LOADTEST_DURATION || "90s";

export const options = smoke
  ? {
      vus: 5,
      duration: "15s",
      thresholds: {
        http_req_failed: ["rate<0.1"],
        checks: ["rate>0.9"],
      },
    }
  : {
      scenarios: {
        slots_list: {
          executor: "constant-arrival-rate",
          rate: rate,
          timeUnit: "1s",
          duration: duration,
          preAllocatedVUs: Math.min(120, Math.max(30, rate * 2)),
          maxVUs: Math.min(400, Math.max(100, rate * 5)),
        },
      },
      thresholds: {
        http_req_failed: ["rate<0.01"],
        http_req_duration: ["p(95)<500", "p(99)<1000"],
        checks: ["rate>0.99"],
      },
    };

export function setup() {
  const res = http.post(
    `${BASE_URL}/api/dummyLogin`,
    JSON.stringify({ role: "user" }),
    { headers: { "Content-Type": "application/json" } },
  );
  if (res.status !== 200) {
    throw new Error(`dummyLogin failed: ${res.status} ${String(res.body).slice(0, 200)}`);
  }
  let body;
  try {
    body = res.json();
  } catch (e) {
    throw new Error(`dummyLogin JSON: ${e}`);
  }
  const token = body && body.token;
  if (!token) {
    throw new Error("dummyLogin: no token in response");
  }
  return {
    token,
    date: nextWeekdayUTCISO(),
  };
}

export default function (data) {
  const url = `${BASE_URL}/api/rooms/${ROOM_ID}/slots/list?date=${data.date}`;
  const res = http.get(url, {
    headers: { Authorization: `Bearer ${data.token}` },
    tags: { name: "SlotsList" },
  });
  const ok = check(res, {
    "status 200": (r) => r.status === 200,
    "json slots": (r) => {
      try {
        const j = r.json();
        return j && Array.isArray(j.slots);
      } catch (_) {
        return false;
      }
    },
  });
  if (!ok && res.status !== 200) {
    console.error(`FAIL ${res.status} ${String(res.body).slice(0, 300)}`);
  }
  if (smoke) {
    sleep(0.3);
  }
}
