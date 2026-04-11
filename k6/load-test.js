// BKL-1056: Teste de carga do Pulse com k6
// Meta: P99 < 200ms em 500 req/s com 1000 usuários simultâneos
//
// Uso:
//   k6 run k6/load-test.js
//   k6 run --out influxdb=http://influxdb:8086/k6 k6/load-test.js  (com InfluxDB/Grafana)
//
// Cenários testados:
//   1. GET /flags — leitura pública de feature flags (cache Redis)
//   2. POST /api/v1/events/track — ingestão de evento individual
//   3. POST /api/v1/events/batch — ingestão em lote (10 eventos)
//   4. GET /flags/evaluate?user_id=... — avaliação personalizada por usuário

import http from "k6/http";
import { check, sleep } from "k6";
import { Rate, Trend } from "k6/metrics";

// ─── Configuração ────────────────────────────────────────────────────────────

const BASE_URL = __ENV.PULSE_URL || "https://pulse-api.institutoitinerante.com.br";
const SERVICE_TOKEN = __ENV.SERVICE_TOKEN || "test-token";
const PROJECT_ID = __ENV.PROJECT_ID || "00000000-0000-0000-0000-000000000001";

// ─── Métricas customizadas ────────────────────────────────────────────────────

const errorRate = new Rate("errors");
const flagsLatency = new Trend("flags_latency", true);
const trackLatency = new Trend("track_latency", true);
const batchLatency = new Trend("batch_latency", true);

// ─── Cenário de carga ─────────────────────────────────────────────────────────

export const options = {
  scenarios: {
    // Ramp-up: 0 → 1000 VUs em 2 minutos, sustentado por 5 minutos
    main_load: {
      executor: "ramping-vus",
      startVUs: 0,
      stages: [
        { duration: "1m", target: 100 },   // warm-up
        { duration: "1m", target: 500 },   // ramp
        { duration: "5m", target: 1000 },  // peak load
        { duration: "1m", target: 0 },     // cool-down
      ],
      gracefulRampDown: "30s",
    },

    // Spike test: burst de 2000 VUs por 30 segundos
    spike_test: {
      executor: "constant-vus",
      vus: 2000,
      duration: "30s",
      startTime: "8m",
    },
  },

  thresholds: {
    // SLOs: P99 < 200ms, error rate < 1%
    http_req_duration: ["p(99)<200", "p(95)<100", "avg<50"],
    errors: ["rate<0.01"],
    flags_latency: ["p(99)<150"],
    track_latency: ["p(99)<250"],
    batch_latency: ["p(99)<300"],
  },

  // Sumário em JSON para integração com CI
  summaryTrendStats: ["avg", "min", "med", "max", "p(90)", "p(95)", "p(99)"],
};

// ─── Dados de teste ───────────────────────────────────────────────────────────

const EVENT_TYPES = [
  "page_view",
  "checkout_started",
  "checkout_completed",
  "button_clicked",
  "feature_used",
];

const PRODUCTS = ["libri", "nitro", "brio", "socialmake"];

function randomItem(arr) {
  return arr[Math.floor(Math.random() * arr.length)];
}

function randomUserId() {
  return `user-${Math.floor(Math.random() * 10000)}`;
}

// ─── Cenário principal ────────────────────────────────────────────────────────

export default function main() {
  const userId = randomUserId();
  const headers = {
    "Content-Type": "application/json",
    Accept: "application/json",
  };
  const authHeaders = {
    ...headers,
    "X-Service-Token": SERVICE_TOKEN,
  };

  // ─ 1. GET /flags — leitura de flags (40% das requisições)
  if (Math.random() < 0.4) {
    const r = http.get(`${BASE_URL}/flags?project_id=${PROJECT_ID}`, { headers });
    flagsLatency.add(r.timings.duration);
    const ok = check(r, {
      "flags: status 200": (res) => res.status === 200,
      "flags: has flags key": (res) => {
        try {
          return JSON.parse(res.body).flags !== undefined;
        } catch {
          return false;
        }
      },
    });
    if (!ok) errorRate.add(1);
    else errorRate.add(0);
  }

  // ─ 2. POST /api/v1/events/track — evento individual (30% das requisições)
  if (Math.random() < 0.3) {
    const payload = JSON.stringify({
      event_name: randomItem(EVENT_TYPES),
      user_id: userId,
      product: randomItem(PRODUCTS),
      environment: "production",
      properties: {
        page: "/test",
        referrer: "k6-load-test",
      },
    });
    const r = http.post(`${BASE_URL}/api/v1/events/track`, payload, {
      headers: authHeaders,
    });
    trackLatency.add(r.timings.duration);
    const ok = check(r, {
      "track: status 201": (res) => res.status === 201,
    });
    if (!ok) errorRate.add(1);
    else errorRate.add(0);
  }

  // ─ 3. POST /api/v1/events/batch — ingestão em lote (20% das requisições)
  if (Math.random() < 0.2) {
    const events = Array.from({ length: 10 }, () => ({
      event_name: randomItem(EVENT_TYPES),
      user_id: randomUserId(),
      product: randomItem(PRODUCTS),
      environment: "production",
    }));
    const payload = JSON.stringify({ events });
    const r = http.post(`${BASE_URL}/api/v1/events/batch`, payload, {
      headers: authHeaders,
    });
    batchLatency.add(r.timings.duration);
    const ok = check(r, {
      "batch: status 2xx": (res) => res.status >= 200 && res.status < 300,
    });
    if (!ok) errorRate.add(1);
    else errorRate.add(0);
  }

  // ─ 4. GET /flags/evaluate — avaliação por user_id (10% das requisições)
  if (Math.random() < 0.1) {
    const r = http.get(
      `${BASE_URL}/flags/evaluate?user_id=${userId}&project_id=${PROJECT_ID}`,
      { headers }
    );
    const ok = check(r, {
      "evaluate: status 200": (res) => res.status === 200,
    });
    if (!ok) errorRate.add(1);
    else errorRate.add(0);
  }

  sleep(Math.random() * 0.1); // think time: 0-100ms
}

// ─── Sumário final ─────────────────────────────────────────────────────────────

export function handleSummary(data) {
  const passed = data.metrics.http_req_duration?.values?.["p(99)"] < 200;
  const errorRateVal = data.metrics.errors?.values?.rate ?? 0;

  return {
    stdout: `
=== PULSE LOAD TEST SUMMARY ===
SLO P99 < 200ms: ${passed ? "✅ PASSED" : "❌ FAILED"}
Error rate: ${(errorRateVal * 100).toFixed(2)}% (limit: 1%)
Flags P99: ${data.metrics.flags_latency?.values?.["p(99)"]?.toFixed(1) ?? "N/A"}ms
Track P99: ${data.metrics.track_latency?.values?.["p(99)"]?.toFixed(1) ?? "N/A"}ms
Batch P99: ${data.metrics.batch_latency?.values?.["p(99)"]?.toFixed(1) ?? "N/A"}ms
Total requests: ${data.metrics.http_reqs?.values?.count ?? 0}
================================
`,
    "k6/results.json": JSON.stringify(data),
  };
}
