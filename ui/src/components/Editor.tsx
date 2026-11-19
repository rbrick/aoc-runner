import type { KeyboardEvent } from "react";
import type { Language } from "../api";

const labels: Record<string, string> = {
  go: "Go",
  java: "Java",
  odin: "Odin",
  python: "Python",
  rust: "Rust",
  typescript: "TypeScript",
};

// go and odin are formatted with tabs, everything else with spaces
const indents: Record<string, string> = {
  go: "\t",
  odin: "\t",
};

type Props = {
  languages: Language[];
  language: string;
  code: string;
  running: boolean;
  onLanguageChange: (language: string) => void;
  onCodeChange: (code: string) => void;
  onRun: () => void;
};

export function Editor({ languages, language, code, running, onLanguageChange, onCodeChange, onRun }: Props) {
  function onKeyDown(event: KeyboardEvent<HTMLTextAreaElement>) {
    if (event.key === "Enter" && (event.metaKey || event.ctrlKey)) {
      event.preventDefault();
      onRun();
      return;
    }

    // tab indents instead of leaving the editor, shift+tab still moves focus
    if (event.key === "Tab" && !event.shiftKey) {
      event.preventDefault();

      const editor = event.currentTarget;
      const indent = indents[language] ?? "    ";
      const { selectionStart, selectionEnd } = editor;

      onCodeChange(code.slice(0, selectionStart) + indent + code.slice(selectionEnd));

      requestAnimationFrame(() => {
        editor.selectionStart = editor.selectionEnd = selectionStart + indent.length;
      });
    }
  }

  return (
    <div className="editor">
      <div className="editor-bar">
        <label className="field field-inline">
          <span className="field-label">Language</span>
          <select value={language} onChange={(event) => onLanguageChange(event.target.value)}>
            {languages.map((language) => (
              <option key={language.name} value={language.name}>
                {labels[language.name] ?? language.name}
              </option>
            ))}
          </select>
        </label>

        <button className="button button-primary" onClick={onRun} disabled={running || !code.trim()}>
          {running ? "Running…" : "Run"}
          <kbd>Ctrl ↵</kbd>
        </button>
      </div>

      <textarea
        className="code"
        value={code}
        onChange={(event) => onCodeChange(event.target.value)}
        onKeyDown={onKeyDown}
        spellCheck={false}
        autoCapitalize="off"
        autoComplete="off"
        aria-label="Solution code"
      />
    </div>
  );
}
