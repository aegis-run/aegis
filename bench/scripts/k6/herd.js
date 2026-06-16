import { check } from "k6";
import { Counter, Rate, Trend } from "k6/metrics";
import { textSummary } from "https://jslib.k6.io/k6-summary/0.0.2/index.js";
import { StatusOK } from "k6/net/grpc";
import * as lib from "./lib.js";

const HERD_PATH = __ENV.HERD_PATH || "../../out/baseline/herd_check.json";
const OUT_DIR = __ENV.OUT_DIR || "out/baseline";
const CONCURRENCY = Number(__ENV.CONCURRENCY || "100");

export const decisionMismatch = new Rate("decision_mismatch");

const cacheHits = new Counter(lib.CACHE_HITS_TOTAL);
const cacheMisses = new Counter(lib.CACHE_MISSES_TOTAL);
const singleflight = new Counter(lib.SINGLEFLIGHT_SHARED);
const dbOps = new Counter(lib.DB_OPS_TOTAL);
const fanoutSum = new Counter(lib.ENGINE_FANOUT_SUM);

const herdCheck = JSON.parse(open(HERD_PATH));

export const options = {
  scenarios: {
    herd: {
      executor: "per-vu-iterations",
      vus: CONCURRENCY,
      iterations: 1,
      maxDuration: "30s",
    },
  },
  thresholds: {
    checks: ["rate>0.99"],
    decision_mismatch: ["rate==0"],
    grpc_req_duration: ["p(95)<500"],
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

  const res = client.invoke(
    "aegis.authz.v1.Authz/Authorize",
    lib.toAuthorizeRequest(herdCheck),
    {
      metadata: { authorization: `Bearer ${lib.TOKEN}` },
      tags: { class: herdCheck.class },
    },
  );

  const okStatus = res && res.status === StatusOK;
  const actual = okStatus ? lib.normalizeDecision(res.message.decision) : "unknown";
  const expected = lib.expectedDecision(herdCheck.expected);
  const matched = actual === expected;

  decisionMismatch.add(!matched, { class: herdCheck.class });

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
  fanoutSum.add(post.fanoutSum - pre.fanoutSum);
}

export function handleSummary(data) {
  const metricCount = (name) => data.metrics[name]?.values?.count ?? 0;

  const shared = metricCount(lib.SINGLEFLIGHT_SHARED);
  // const ops = metricCount(lib.DB_OPS_TOTAL);
  const iterations = metricCount("iterations");

  const coalescingRate = iterations > 0 ? shared / iterations : 0;

  // Snapshot the raw data object before any mutations to keep the JSON report pristine
  const rawJSON = JSON.stringify(data);

  // Inject the computed coalescing rate into data.metrics so textSummary formats it natively
  data.metrics.aegis_coalescing_rate = {
    type: "gauge",
    contains: "default",
    values: { value: coalescingRate, min: coalescingRate, max: coalescingRate },
  };

  const reportPath = `${OUT_DIR}/results/herd_${CONCURRENCY}.json`;

  return {
    stdout: textSummary(data, { indent: " ", enableColors: true }),
    [reportPath]: rawJSON,
  };
}
