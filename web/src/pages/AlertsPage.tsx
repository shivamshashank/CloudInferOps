import { useState } from "react";
import { Alert } from "../types";
import Status from "../components/Status";
import Empty from "../components/Empty";
import { tone, when } from "../utils";

export default function AlertsPage({ items }: { items: Alert[] }) {
  const [filter, setFilter] = useState("all");
  const filtered = items.filter(
    (i) => filter === "all" || i.status.toLowerCase() === filter,
  );
  return (
    <section className="panel">
      <div className="toolbar">
        <div className="tabs">
          {["all", "active", "resolved"].map((x) => (
            <button
              key={x}
              className={filter === x ? "active" : ""}
              onClick={() => setFilter(x)}
            >
              {x}
            </button>
          ))}
        </div>
        <span>{filtered.length} incidents</span>
      </div>
      {filtered.map((item) => (
        <div className="alert-row" key={`${item.title}/${item.timestamp}`}>
          <div className={`severity ${tone(item.severity)}`}>△</div>
          <div>
            <strong>{item.title}</strong>
            <small>{when(item.timestamp)}</small>
          </div>
          <Status value={item.severity} />
          <Status value={item.status} />
        </div>
      ))}
      {filtered.length === 0 ? (
        <Empty
          title="No incidents in this view"
          copy="Alertmanager events will appear here through the webhook handler."
        />
      ) : null}
    </section>
  );
}
