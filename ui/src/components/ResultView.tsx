import { useState } from "react";
import type { Result, SubmitResult } from "../api";

type SubmitPart = (level: 1 | 2, answer: string) => Promise<SubmitResult>;

type Props = {
  result: Result | null;
  error: string | null;
  // the puzzle's example answers, when an example was used as input
  expected: string[] | null;
  // set when the answers can be submitted to AoC, only for your own input
  submitPart: SubmitPart | null;
  onSolved: () => void;
};

export function ResultView({ result, error, expected, submitPart, onSolved }: Props) {
  if (error) {
    return (
      <section className="result">
        <Status kind="error" label="Request failed" />
        <pre className="output">{error}</pre>
      </section>
    );
  }

  if (!result) {
    return (
      <section className="result result-empty">
        <p className="muted">Run your solution to see the answers here.</p>
      </section>
    );
  }

  const { build, run } = result;

  if (!run) {
    return (
      <section className="result">
        <Status kind="error" label="Compile error" detail={`${build.duration_ms} ms`} />
        <pre className="output">{(build.errors ?? []).join("\n")}</pre>
      </section>
    );
  }

  const failed = !result.ok;

  return (
    <section className="result">
      <Status
        kind={failed ? "error" : "ok"}
        label={failed ? "Runtime error" : "Finished"}
        detail={`build ${build.duration_ms} ms · run ${run.duration_ms} ms`}
      />

      <div className="parts">
        <Answer
          part="Part one"
          level={1}
          value={run.part_one}
          expected={expected?.[0]}
          submitPart={submitPart}
          onSolved={onSolved}
        />
        <Answer
          part="Part two"
          level={2}
          value={run.part_two}
          expected={expected?.[1]}
          submitPart={submitPart}
          onSolved={onSolved}
        />
      </div>

      {run.errors && run.errors.length > 0 && (
        <>
          <h3 className="output-title">Output</h3>
          <pre className="output">{run.errors.join("\n")}</pre>
        </>
      )}
    </section>
  );
}

function Status({ kind, label, detail }: { kind: "ok" | "error"; label: string; detail?: string }) {
  return (
    <div className="status">
      <span className={kind === "ok" ? "pill pill-ok" : "pill pill-error"}>{label}</span>
      {detail && <span className="muted">{detail}</span>}
    </div>
  );
}

type AnswerProps = {
  part: string;
  level: 1 | 2;
  value: string;
  expected?: string;
  submitPart: SubmitPart | null;
  onSolved: () => void;
};

function Answer({ part, level, value, expected, submitPart, onSolved }: AnswerProps) {
  // compare ignoring surrounding whitespace, solutions often return a trailing newline
  const matches = expected ? value.trim() === expected.trim() : null;

  let className = "answer";

  if (matches !== null) {
    className = matches ? "answer answer-match" : "answer answer-mismatch";
  }

  return (
    <div className={className}>
      <span className="answer-label">{part}</span>
      {value ? <pre className="answer-value">{value}</pre> : <span className="muted">no answer</span>}
      {matches !== null && (
        <span className="answer-expected">
          {matches ? (
            "✓ matches the example answer"
          ) : (
            <>
              ✗ example answer is <code>{expected}</code>
            </>
          )}
        </span>
      )}
      {submitPart && value.trim() && (
        <SubmitAnswer level={level} part={part} answer={value.trim()} submitPart={submitPart} onSolved={onSolved} />
      )}
    </div>
  );
}

type SubmitState =
  | { status: "idle" }
  | { status: "confirming" }
  | { status: "submitting" }
  | { status: "done"; result: SubmitResult }
  | { status: "error"; message: string };

const outcomeLabels: Record<SubmitResult["outcome"], string> = {
  correct: "★ Correct",
  incorrect: "✗ Not the right answer",
  rate_limited: "Too soon",
  wrong_level: "Not accepted",
  unknown: "Unexpected response",
};

function formatWait(seconds: number) {
  const minutes = Math.floor(seconds / 60);
  return minutes > 0 ? `${minutes}m ${seconds % 60}s` : `${seconds}s`;
}

function SubmitAnswer({
  level,
  part,
  answer,
  submitPart,
  onSolved,
}: {
  level: 1 | 2;
  part: string;
  answer: string;
  submitPart: SubmitPart;
  onSolved: () => void;
}) {
  const [state, setState] = useState<SubmitState>({ status: "idle" });

  async function send() {
    setState({ status: "submitting" });

    try {
      const result = await submitPart(level, answer);
      setState({ status: "done", result });

      if (result.outcome === "correct") {
        onSolved();
      }
    } catch (err) {
      setState({ status: "error", message: (err as Error).message });
    }
  }

  // a wrong answer costs a timeout on AoC, so submitting always takes a second click
  if (state.status === "idle") {
    return (
      <div className="submit">
        <button className="button button-small" onClick={() => setState({ status: "confirming" })}>
          Submit to AoC
        </button>
      </div>
    );
  }

  if (state.status === "confirming") {
    return (
      <div className="submit submit-confirm">
        <span>
          Submit <code>{answer}</code> as your {part.toLowerCase()} answer?
        </span>
        <span className="submit-actions">
          <button className="button button-small button-primary" onClick={send}>
            Submit
          </button>
          <button className="button button-small" onClick={() => setState({ status: "idle" })}>
            Cancel
          </button>
        </span>
      </div>
    );
  }

  if (state.status === "submitting") {
    return <div className="submit muted">Submitting…</div>;
  }

  if (state.status === "error") {
    return (
      <div className="submit submit-incorrect">
        <span>{state.message}</span>
        <button className="button button-small" onClick={() => setState({ status: "idle" })}>
          Try again
        </button>
      </div>
    );
  }

  const { result } = state;
  const retry = result.outcome === "rate_limited" || result.outcome === "unknown";

  return (
    <div className={`submit submit-${result.outcome === "correct" ? "correct" : "incorrect"}`}>
      <strong>{outcomeLabels[result.outcome]}</strong>
      {result.outcome === "rate_limited" && result.wait_seconds ? (
        <span>Wait {formatWait(result.wait_seconds)} before submitting again.</span>
      ) : (
        <span>{result.message}</span>
      )}
      {result.duplicate && <span className="muted">You already submitted this answer, so it wasn't sent again.</span>}
      {retry && (
        <button className="button button-small" onClick={() => setState({ status: "idle" })}>
          Try again
        </button>
      )}
    </div>
  );
}
