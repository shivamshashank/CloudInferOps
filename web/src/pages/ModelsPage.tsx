import { useMemo, useState } from "react";
import { Model } from "../types";
import Status from "../components/Status";
import Empty from "../components/Empty";

export default function ModelsPage({ items }: { items: Model[] }) {
  const [query, setQuery] = useState("");
  const filtered = useMemo(
    () =>
      items.filter((i) =>
        `${i.name} ${i.provider}`.toLowerCase().includes(query.toLowerCase()),
      ),
    [items, query],
  );
  return (
    <section className="panel">
      <div className="toolbar">
        <input
          aria-label="Search models"
          placeholder="Search model inventory…"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
        />
        <span>{filtered.length} models</span>
      </div>
      <div className="card-grid">
        {filtered.map((item) => (
          <article className="model-card" key={item.name}>
            <div className="model-icon">◇</div>
            <Status value={item.status} />
            <h3>{item.name}</h3>
            <p>{item.provider} provider</p>
            <footer>
              <span>{item.location}</span>
              <span>OpenAI compatible</span>
            </footer>
          </article>
        ))}
      </div>
      {filtered.length === 0 ? (
        <Empty
          title="No models discovered"
          copy="Deploy Ollama or vLLM and make sure the gateway /models endpoint is reachable."
        />
      ) : null}
    </section>
  );
}
