import { useState } from "react";
import type { Problem } from "../api";

type Props = {
  problem: Problem | null;
  error: string | null;
  onUseExample: (index: number) => void;
};

export function ProblemView({ problem, error, onUseExample }: Props) {
  if (error) {
    return <p className="muted">Couldn't load the puzzle: {error}</p>;
  }

  if (!problem) {
    return <p className="muted">Loading puzzle…</p>;
  }

  const examples = problem.examples ?? [];
  const answers = problem.example_answers ?? [];

  return (
    <div className="problem">
      {/* the puzzle text is HTML fetched from adventofcode.com by the backend */}
      <div className="problem-body" dangerouslySetInnerHTML={{ __html: problem.body }} />

      {examples.length > 0 && (
        <section className="examples">
          <h3>Examples</h3>
          {examples.map((example, index) => (
            <Example key={index} index={index} text={example} onUse={() => onUseExample(index)} />
          ))}

          <h3>Example answers</h3>
          <dl className="answers">
            {answers.map((answer, index) => (
              <div key={index}>
                <dt>Part {index === 0 ? "one" : "two"}</dt>
                <dd>{answer ? <code>{answer}</code> : <span className="muted">not found</span>}</dd>
              </div>
            ))}
          </dl>
        </section>
      )}
    </div>
  );
}

function Example({ index, text, onUse }: { index: number; text: string; onUse: () => void }) {
  const [copied, setCopied] = useState(false);

  async function copy() {
    try {
      await navigator.clipboard.writeText(text);
      setCopied(true);
      setTimeout(() => setCopied(false), 1500);
    } catch {
      // clipboard access was denied, nothing else to do
    }
  }

  return (
    <figure className="example">
      <figcaption>
        <span>Example {index + 1}</span>
        <span className="example-actions">
          <button className="button button-small" onClick={onUse}>
            Use as input
          </button>
          <button className="button button-small" onClick={copy}>
            {copied ? "Copied" : "Copy"}
          </button>
        </span>
      </figcaption>
      <pre>{text}</pre>
    </figure>
  );
}
