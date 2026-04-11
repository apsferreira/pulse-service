// BKL-1056: Smoke test do Pulse com k6 — verifica endpoints básicos antes de load test.
// Uso: k6 run k6/smoke-test.js

import http from "k6/http";
import { check } from "k6";

const BASE_URL = __ENV.PULSE_URL || "https://pulse-api.institutoitinerante.com.br";

export const options = {
  vus: 1,
  iterations: 10,
  thresholds: {
    http_req_duration: ["p(99)<500"],
    http_req_failed: ["rate<0.01"],
  },
};

export default function () {
  // Health check
  const health = http.get(`${BASE_URL}/health`);
  check(health, {
    "health: status 200": (r) => r.status === 200,
    "health: service is pulse-service": (r) => {
      try { return JSON.parse(r.body).service === "pulse-service"; } catch { return false; }
    },
  });

  // GET /flags
  const flags = http.get(`${BASE_URL}/flags`);
  check(flags, {
    "flags: status 200": (r) => r.status === 200,
    "flags: has flags key": (r) => {
      try { return "flags" in JSON.parse(r.body); } catch { return false; }
    },
  });

  // GET /metrics (Prometheus format)
  const metrics = http.get(`${BASE_URL}/metrics`);
  check(metrics, {
    "metrics: status 200": (r) => r.status === 200,
    "metrics: has pulse_uptime": (r) => r.body.includes("pulse_uptime_seconds"),
  });
}
