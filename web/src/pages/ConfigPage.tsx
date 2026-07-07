import { useEffect, useState } from "react";
import { PortalConfig } from "../types";

export default function ConfigPage({
  value,
  actions,
  onSave,
}: {
  value: PortalConfig;
  actions: boolean;
  onSave: (v: PortalConfig) => void;
}) {
  const [form, setForm] = useState(value);
  useEffect(() => setForm(value), [value]);
  return (
    <section className="panel config">
      <div>
        <span className="kicker">SAFE CONFIGURATION</span>
        <h3>Inference defaults</h3>
        <p>
          Updates the model-config ConfigMap. Secrets are deliberately excluded
          from this portal.
        </p>
      </div>
      <div className="config-grid">
        <label>
          Operational namespace
          <input
            value={form.namespace}
            onChange={(e) => setForm({ ...form, namespace: e.target.value })}
          />
        </label>
        <label>
          Provider
          <select
            value={form.provider}
            onChange={(e) => setForm({ ...form, provider: e.target.value })}
          >
            <option value="ollama">Ollama</option>
            <option value="vllm">vLLM</option>
          </select>
        </label>
        <label>
          Default model
          <input
            value={form.model}
            onChange={(e) => setForm({ ...form, model: e.target.value })}
          />
        </label>
      </div>
      <button disabled={!actions} onClick={() => onSave(form)}>
        Save configuration
      </button>
      <p className="guard">
        {actions
          ? "Changes are applied directly to the in-cluster ConfigMap."
          : "Configuration is read-only while actions are disabled."}
      </p>
    </section>
  );
}
