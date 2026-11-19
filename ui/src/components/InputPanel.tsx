export type InputSource = number | "mine" | "custom";

export type MyInput =
  | { status: "idle" }
  | { status: "loading" }
  | { status: "ready"; input: string }
  | { status: "error"; message: string };

type Props = {
  examples: string[];
  source: InputSource;
  value: string;
  myInput: MyInput;
  onSourceChange: (source: InputSource) => void;
  onCustomChange: (input: string) => void;
};

function parseSource(value: string): InputSource {
  if (value === "custom" || value === "mine") {
    return value;
  }

  return Number(value);
}

export function InputPanel({ examples, source, value, myInput, onSourceChange, onCustomChange }: Props) {
  const lines = value.split("\n").filter(Boolean).length;

  return (
    <div className="input-panel">
      <div className="input-bar">
        <label className="field field-inline">
          <span className="field-label">Input</span>
          <select value={String(source)} onChange={(event) => onSourceChange(parseSource(event.target.value))}>
            {examples.map((_, index) => (
              <option key={index} value={index}>
                Example {index + 1}
              </option>
            ))}
            <option value="mine">Your input</option>
            <option value="custom">Custom</option>
          </select>
        </label>
        <span className="muted input-size">{value ? `${lines} ${lines === 1 ? "line" : "lines"}` : "empty"}</span>
      </div>

      {source === "mine" && myInput.status === "loading" && <p className="muted input-status">Loading your input…</p>}
      {source === "mine" && myInput.status === "error" && (
        <p className="input-status input-error">{myInput.message}</p>
      )}

      {/* editing an example or your input turns it into custom input */}
      <textarea
        className="code input"
        value={value}
        onChange={(event) => onCustomChange(event.target.value)}
        placeholder="Paste input here, it's passed to your solution on stdin"
        spellCheck={false}
        autoCapitalize="off"
        autoComplete="off"
        aria-label="Input"
      />
    </div>
  );
}
