import MonacoEditor, { type OnMount } from "@monaco-editor/react";
import { useEffect, useRef, useState } from "react";
import type { Language } from "../api";
import "../monaco";

const labels: Record<string, string> = {
  go: "Go",
  haskell: "Haskell",
  java: "Java",
  odin: "Odin",
  python: "Python",
  rust: "Rust",
  typescript: "TypeScript",
};

// go and odin are formatted with tabs, everything else with spaces
const usesTabs = new Set(["go", "odin"]);

function useDarkMode() {
  const query = "(prefers-color-scheme: dark)";
  const [dark, setDark] = useState(() => window.matchMedia(query).matches);

  useEffect(() => {
    const media = window.matchMedia(query);
    const onChange = () => setDark(media.matches);

    media.addEventListener("change", onChange);
    return () => media.removeEventListener("change", onChange);
  }, []);

  return dark;
}

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
  const dark = useDarkMode();

  // the keybinding is registered once, so it has to call the latest onRun
  const runRef = useRef(onRun);
  runRef.current = onRun;

  const onMount: OnMount = (editor, monaco) => {
    editor.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.Enter, () => runRef.current());
  };

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

      <div className="monaco">
        <div className="monaco-inner">
          <MonacoEditor
            height="100%"
            language={language}
            value={code}
            theme={dark ? "aoc-runner-dark" : "aoc-runner-light"}
            onChange={(value) => onCodeChange(value ?? "")}
            onMount={onMount}
            loading={<p className="muted">Loading editor…</p>}
            options={{
              fontFamily: 'ui-monospace, "SF Mono", "JetBrains Mono", Menlo, Consolas, monospace',
              fontSize: 13.5,
              lineHeight: 22,
              tabSize: 4,
              insertSpaces: !usesTabs.has(language),
              detectIndentation: false,
              minimap: { enabled: false },
              scrollBeyondLastLine: false,
              automaticLayout: true,
              padding: { top: 12, bottom: 12 },
              renderLineHighlight: "line",
              ariaLabel: "Solution code",
            }}
          />
        </div>
      </div>
    </div>
  );
}
