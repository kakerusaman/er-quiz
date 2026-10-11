import { useState } from "react";

// 業務の説明文。見出しかリンクを押すと開閉する。最初は開いた状態。
export function ScenarioPanel({ scenario }: { scenario: string }) {
  const [open, setOpen] = useState(true);
  const toggle = () => setOpen((o) => !o);
  return (
    <section className="card scenario">
      <div className="scenario-head">
        <button className="scenario-title" aria-expanded={open} aria-controls="scenario-body" onClick={toggle}>
          <span className="caret" aria-hidden="true">
            {open ? "▼" : "▶"}
          </span>
          業務の説明文
        </button>
        <button className="link small" onClick={toggle}>
          {open ? "小さくする" : "広げる"}
        </button>
      </div>
      <p id="scenario-body" className="scenario-body" hidden={!open}>
        {scenario}
      </p>
    </section>
  );
}
