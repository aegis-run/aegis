import http from "k6/http";
import grpc from "k6/net/grpc";

export const GRPC_ADDR = __ENV.GRPC_ADDR || "localhost:43615";
export const TOKEN =
  __ENV.AEGIS_TOKEN || "aegis_XBHFNLxp9usfmbAuZtrD0Ajiace1HNfb/CpVLy3qcek=";

export function newClient() {
  const client = new grpc.Client();
  return client;
}

export function connect(client) {
  client.connect(GRPC_ADDR, {
    plaintext: true,
    reflect: true,
  });
}

export function toAuthorizeRequest(c) {
  return {
    resource: {
      type: c.resource.type,
      id: c.resource.id,
    },
    permission: c.permission,
    actor: {
      type: c.actor.type,
      id: c.actor.id,
    },
  };
}

export function normalizeDecision(value) {
  // Handle numeric values (1=Allowed, 2=Denied)
  if (value === 1 || value === "1") return "allowed";
  if (value === 2 || value === "2") return "denied";

  // Handle string values (case-insensitive and partial match)
  const s = String(value).toLowerCase();
  if (s.includes("allowed")) return "allowed";
  if (s.includes("denied")) return "denied";

  return "unknown";
}

export function expectedDecision(expected) {
  return expected ? "allowed" : "denied";
}

export const METRICS_URL = __ENV.METRICS_URL || "http://localhost:43614/metrics";

export const CACHE_HITS_TOTAL = "aegis_dispatch_cache_hits_total";
export const CACHE_MISSES_TOTAL = "aegis_dispatch_cache_misses_total";
export const SINGLEFLIGHT_SHARED = "aegis_dispatch_singleflight_shared_total";
export const DB_OPS_TOTAL = "aegis_database_operations_total";
export const TUPLES_FETCHED_TOTAL = "aegis_engine_tuples_fetched_total";
export const DB_LATENCY_SUM = "aegis_database_operations_latency_milliseconds_sum";
export const ENGINE_FANOUT_SUM = "aegis_engine_fanout_size_sum";
export const ENGINE_FANOUT_COUNT = "aegis_engine_fanout_size_count";

export function scrapeMetrics() {
  const res = http.get(METRICS_URL);

  return {
    hits: getMetric(res.body, CACHE_HITS_TOTAL),
    misses: getMetric(res.body, CACHE_MISSES_TOTAL),
    shared: getMetric(res.body, SINGLEFLIGHT_SHARED),
    dbOps: getMetric(res.body, DB_OPS_TOTAL),
    tuples: getMetric(res.body, TUPLES_FETCHED_TOTAL),
    dbLatency: getMetric(res.body, DB_LATENCY_SUM),
    fanoutSum: getMetric(res.body, ENGINE_FANOUT_SUM),
    fanoutCount: getMetric(res.body, ENGINE_FANOUT_COUNT),
  };
}

function getMetric(body, metric) {
  // Matches the metric at the start of a line, ignores labels, and captures the number
  const regex = new RegExp(`^${metric}(?:\\{[^}]*\\})?\\s+([0-9.eE+-]+)`, "gm");
  let sum = 0;
  let match;
  while ((match = regex.exec(body)) !== null) {
    sum += parseFloat(match[1]);
  }
  return sum;
}
