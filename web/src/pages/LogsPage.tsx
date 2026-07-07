import { useEffect, useState } from "react";
import { Pod } from "../types";
import Empty from "../components/Empty";
import { api, tone } from "../utils";

export default function LogsPage({ pods }: { pods: Pod[] }) {
  const [selected, setSelected] = useState("");
  const [lines, setLines] = useState<string[]>([]);
  const [query, setQuery] = useState("");
  const [busy, setBusy] = useState(false);
  const [controller, setController] = useState<AbortController | null>(null);
  const chosen = pods.find((p) => `${p.namespace}/${p.name}` === selected);

  useEffect(() => () => controller?.abort(), [controller]);

  const load = async () => {
    if (!chosen) return;
    setBusy(true);
    try {
      const data = await api<{ lines: string[] }>(
        `logs?namespace=${encodeURIComponent(
          chosen.namespace,
        )}&pod=${encodeURIComponent(chosen.name)}&lines=200`,
      );
      setLines(data.lines);
    } finally {
      setBusy(false);
    }
  };

  const stream = async () => {
    if (!chosen) return;
    if (controller) {
      controller.abort();
      setController(null);
      return;
    }
    const next = new AbortController();
    setController(next);
    setLines([]);
    const token = sessionStorage.getItem("cloudinferops-token");
    try {
      const response = await fetch(
        `api/logs/stream?namespace=${encodeURIComponent(
          chosen.namespace,
        )}&pod=${encodeURIComponent(chosen.name)}`,
        {
          signal: next.signal,
          headers: token ? { Authorization: `Bearer ${token}` } : {},
        },
      );
      const reader = response.body?.getReader();
      const decoder = new TextDecoder();
      while (reader) {
        const { done, value } = await reader.read();
        if (done) break;
        const chunk = decoder.decode(value, { stream: true });
        setLines((previous) =>
          [...previous, ...chunk.split("\n").filter(Boolean)].slice(-500),
        );
      }
    } catch (e) {
      if (!next.signal.aborted)
        setLines((previous) => [...previous, `Stream ended: ${String(e)}`]);
    } finally {
      setController(null);
    }
  };

  const shown = lines.filter((line) =>
    line.toLowerCase().includes(query.toLowerCase()),
  );

  return (
    <section className="logs-layout">
      <aside className="panel pod-list">
        <h3>Pods</h3>
        <div className="pod-items">
          {pods.map((p) => (
            <button
              className={
                selected === `${p.namespace}/${p.name}` ? "active" : ""
              }
              key={`${p.namespace}/${p.name}`}
              onClick={() => setSelected(`${p.namespace}/${p.name}`)}
            >
              <i className={tone(p.status)} />
              <span>
                <strong>{p.name}</strong>
                <small>{p.namespace}</small>
              </span>
            </button>
          ))}
          {pods.length === 0 ? (
            <Empty title="No pods" copy="Cluster pod inventory is empty." />
          ) : null}
        </div>
      </aside>
      <div className="panel console">
        <div className="toolbar">
          <input
            aria-label="Filter logs"
            placeholder="Filter log lines…"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
          />
          <div className="button-group">
            <button
              className="secondary"
              disabled={!chosen || busy}
              onClick={() => void load()}
            >
              {busy ? "Loading…" : "Load latest"}
            </button>
            <button disabled={!chosen} onClick={() => void stream()}>
              {controller ? "Stop stream" : "Stream live"}
            </button>
          </div>
        </div>
        <pre>
          {shown.length
            ? shown.map((line, i) => (
                <span key={i}>
                  <b>{String(i + 1).padStart(3, "0")}</b>
                  {line}
                </span>
              ))
            : "Select a pod and load or stream its logs."}
        </pre>
      </div>
    </section>
  );
}
