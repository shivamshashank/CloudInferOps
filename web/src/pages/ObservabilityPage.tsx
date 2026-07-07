import { Component } from "../types";
import Status from "../components/Status";
import { tone } from "../utils";

export default function ObservabilityPage({ items }: { items: Component[] }) {
  return (
    <>
      <section className="section-intro">
        <div>
          <h3>Telemetry stack</h3>
          <p>Live discovery across the observability namespace.</p>
        </div>
        <Status
          value={`${items.filter((i) => i.status === "Running").length}/${
            items.length
          } running`}
        />
      </section>
      <section className="card-grid">
        {items.map((item) => (
          <article className="component-card" key={item.name}>
            <div className="component-top">
              <div className="model-icon">◉</div>
              <Status value={item.status} />
            </div>
            <h3>{item.name}</h3>
            <p>{item.namespace}</p>
            <div className="signal-line">
              <i className={tone(item.status)} />
              <span>
                {item.status === "Running"
                  ? "Metrics and health signals available"
                  : "Waiting for a ready deployment"}
              </span>
            </div>
          </article>
        ))}
      </section>
    </>
  );
}
