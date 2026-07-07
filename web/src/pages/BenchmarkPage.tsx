import { useState } from "react";
import { Benchmark } from "../types";
import Empty from "../components/Empty";
import { when } from "../utils";

export default function BenchmarkPage({
  result,
  actions,
  model,
  onRun,
}: {
  result: Benchmark | null;
  actions: boolean;
  model: string;
  onRun: (m: string, n: number) => void;
}) {
  const [target, setTarget] = useState(model);
  const [requests, setRequests] = useState(5);
  return (
    <>
      <section className="panel benchmark-form">
        <div>
          <span className="kicker">INFERENCE PERFORMANCE</span>
          <h3>Run a lightweight readiness benchmark</h3>
          <p>
            Send controlled chat requests through the in-cluster gateway and
            measure end-to-end latency.
          </p>
        </div>
        <div className="action-form">
          <label>
            Model
            <input value={target} onChange={(e) => setTarget(e.target.value)} />
          </label>
          <label>
            Requests
            <input
              type="number"
              min="1"
              max="20"
              value={requests}
              onChange={(e) => setRequests(Number(e.target.value))}
            />
          </label>
          <button disabled={!actions} onClick={() => onRun(target, requests)}>
            Run benchmark
          </button>
        </div>
      </section>
      {result ? (
        <section className="metric-grid benchmark-results">
          <article>
            <span>Average latency</span>
            <strong>{result.average_latency_ms.toFixed(0)} ms</strong>
          </article>
          <article>
            <span>Successful</span>
            <strong>
              {result.succeeded}/{result.requests}
            </strong>
          </article>
          <article>
            <span>Failed</span>
            <strong>{result.failed}</strong>
          </article>
          <article>
            <span>Completed</span>
            <strong className="small-value">{when(result.completed_at)}</strong>
          </article>
        </section>
      ) : (
        <Empty
          title="No benchmark result yet"
          copy="Enable actions and run a benchmark to establish a baseline."
        />
      )}
    </>
  );
}
