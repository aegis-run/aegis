import { textSummary } from "https://jslib.k6.io/k6-summary/0.0.2/index.js";
import { check } from "k6";
import { SharedArray } from "k6/data";
import { Counter, Rate, Trend } from "k6/metrics";
import { StatusOK } from "k6/net/grpc";
import * as lib from "./lib.js";

const CHECKS_PATH = __ENV.CHECKS_PATH || "../../dataset/baseline/checks.jsonl";
const OUT_DIR = __ENV.OUT_DIR || "out/baseline";
const RPS = Number(__ENV.RPS || "100");
const DURATION = __ENV.DURATION || "60s";
const PRE_ALLOCATED_VUS = Number(__ENV.PRE_ALLOCATED_VUS || "200");
const MAX_VUS = Number(__ENV.MAX_VUS || "3000");

export const decisionMismatch = new Rate("decision_mismatch");

const traverseLatency = new Trend("latency_traversed", true);
const directLatency = new Trend("latency_direct", true);
const cacheHits = new Counter(lib.CACHE_HITS_TOTAL);
const cacheMisses = new Counter(lib.CACHE_MISSES_TOTAL);
const singleflight = new Counter(lib.SINGLEFLIGHT_SHARED);
const dbOps = new Counter(lib.DB_OPS_TOTAL);
const tuplesFetched = new Counter(lib.TUPLES_FETCHED_TOTAL);
const dbLatencySum = new Trend(lib.DB_LATENCY_SUM, true);
const fanoutSum = new Counter(lib.ENGINE_FANOUT_SUM);

const checks = new SharedArray("checks", () => {
  return open(CHECKS_PATH).trim().split("\n").filter(Boolean).map(JSON.parse);
});

export const options = {
  scenarios: {
    steady: {
      executor: "constant-arrival-rate",
      rate: RPS,
      timeUnit: "1s",
      duration: DURATION,
      preAllocatedVUs: PRE_ALLOCATED_VUS,
      maxVUs: MAX_VUS,
    },
  },
  thresholds: {
    checks: ["rate>0.99"],
    decision_mismatch: ["rate==0"],
    "grpc_req_duration{class:doc_owner_view}": ["p(95)<100"], // Direct access should be fast
  },
};

const client = lib.newClient();

export function setup() {
  return lib.scrapeMetrics();
}

export default function () {
  if (__ITER === 0) {
    lib.connect(client);
  }

  const c = checks[Math.floor(Math.random() * checks.length)];

  const res = client.invoke("aegis.authz.v1.Authz/Authorize", lib.toAuthorizeRequest(c), {
    metadata: {
      authorization: `Bearer ${lib.TOKEN}`,
    },
    tags: { class: c.class },
  });

  const start = Date.now();
  const okStatus = res && res.status === StatusOK;
  const end = Date.now();

  const duration = end - start;
  if (c.class.includes("traversal") || c.class.includes("viewer")) {
    traverseLatency.add(duration);
  } else {
    directLatency.add(duration);
  }

  if (!okStatus) {
    console.error(
      `Request failed: status=${res ? res.status : "N/A"} (type: ${typeof (res ? res.status : "")}), error=${res ? res.error : "N/A"}`,
    );
  }

  const actual = okStatus ? lib.normalizeDecision(res.message.decision) : "unknown";
  const expected = lib.expectedDecision(c.expected);
  const matched = actual === expected;

  if (okStatus && !matched) {
    console.warn(
      `Decision mismatch for class ${c.class}: expected ${expected}, got ${actual} (raw decision: ${res.message.decision})`,
    );
  }

  decisionMismatch.add(!matched, { class: c.class });

  check(res, {
    "status is OK": () => okStatus,
    "decision matches expected": () => matched,
  });
}

export function teardown(pre) {
  const post = lib.scrapeMetrics();

  cacheHits.add(post.hits - pre.hits);
  cacheMisses.add(post.misses - pre.misses);
  singleflight.add(post.shared - pre.shared);
  dbOps.add(post.dbOps - pre.dbOps);
  tuplesFetched.add(post.tuples - pre.tuples);
  dbLatencySum.add(post.dbLatency - pre.dbLatency);
  fanoutSum.add(post.fanoutSum - pre.fanoutSum);
}

export function handleSummary(data) {
  const metricCount = (name) => data.metrics[name]?.values?.count ?? 0;

  const hits = metricCount(lib.CACHE_HITS_TOTAL);
  const misses = metricCount(lib.CACHE_MISSES_TOTAL);
  const ops = metricCount(lib.DB_OPS_TOTAL);
  const fanout = metricCount(lib.ENGINE_FANOUT_SUM);
  const tuples = metricCount(lib.TUPLES_FETCHED_TOTAL);
  const iterations = metricCount("iterations");

  const totalCache = hits + misses;
  const hitRatio = totalCache > 0 ? hits / totalCache : 0;
  const avgFanout = iterations > 0 ? fanout / iterations : 0;
  const avgDbOps = iterations > 0 ? ops / iterations : 0;
  const efficiency = avgDbOps > 0 ? avgFanout / avgDbOps : avgFanout;

  // Snapshot the raw data object before mutation so the JSON report remains pristine
  const rawJSON = JSON.stringify(data);

  // Inject calculated summaries into data.metrics so textSummary formats them natively
  data.metrics.aegis_cache_hit_ratio = {
    type: "gauge",
    contains: "default",
    values: { value: hitRatio, min: hitRatio, max: hitRatio },
  };
  data.metrics.aegis_avg_fanout = {
    type: "gauge",
    contains: "default",
    values: { value: avgFanout, min: avgFanout, max: avgFanout },
  };
  data.metrics.aegis_avg_db_ops = {
    type: "gauge",
    contains: "default",
    values: { value: avgDbOps, min: avgDbOps, max: avgDbOps },
  };
  data.metrics.aegis_work_avoidance_factor = {
    type: "gauge",
    contains: "default",
    values: { value: efficiency, min: efficiency, max: efficiency },
  };
  data.metrics.aegis_avg_tuples_fetched = {
    type: "gauge",
    contains: "default",
    values: {
      value: iterations > 0 ? tuples / iterations : 0,
      min: iterations > 0 ? tuples / iterations : 0,
      max: iterations > 0 ? tuples / iterations : 0,
    },
  };

  const summary = textSummary(data, { indent: " ", enableColors: true });
  const reportPath = `${OUT_DIR}/results/checks_${RPS}.json`;

  return {
    stdout: summary,
    [reportPath]: rawJSON,
  };
}
