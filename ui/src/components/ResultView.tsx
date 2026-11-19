import type { Result } from "../api";

type Props = {
  result: Result | null;
  error: string | null;
};

export function ResultView({ result, error }: Props) {
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
        <Answer part="Part one" value={run.part_one} />
        <Answer part="Part two" value={run.part_two} />
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

function Answer({ part, value }: { part: string; value: string }) {
  return (
    <div className="answer">
      <span className="answer-label">{part}</span>
      {value ? <pre className="answer-value">{value}</pre> : <span className="muted">no answer</span>}
    </div>
  );
}
