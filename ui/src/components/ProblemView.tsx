import { useEffect, useState } from "react";
import { getProblem, type Problem } from "../api";

type Props = {
  year: string;
  day: number;
  // bumped after logging in, so part two shows up for solved puzzles
  refresh: number;
};

export function ProblemView({ year, day, refresh }: Props) {
  const [problem, setProblem] = useState<Problem | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;

    setProblem(null);
    setError(null);

    getProblem(year, day)
      .then((problem) => !cancelled && setProblem(problem))
      .catch((err: Error) => !cancelled && setError(err.message));

    return () => {
      cancelled = true;
    };
  }, [year, day, refresh]);

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
            <Example key={index} index={index} text={example} />
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

function Example({ index, text }: { index: number; text: string }) {
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
        <button className="button button-small" onClick={copy}>
          {copied ? "Copied" : "Copy"}
        </button>
      </figcaption>
      <pre>{text}</pre>
    </figure>
  );
}
