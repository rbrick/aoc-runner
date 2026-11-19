import { useEffect, useState } from "react";
import { getLanguages, submit, type Language, type Result } from "./api";
import { AuthStatus } from "./components/AuthStatus";
import { Editor } from "./components/Editor";
import { ProblemView } from "./components/ProblemView";
import { PuzzleBrowser } from "./components/PuzzleBrowser";
import { ResultView } from "./components/ResultView";
import { load, save } from "./storage";

// code is remembered per puzzle and language
function codeKey(year: string | null, day: number | null, language: string) {
  const puzzle = year && day ? `${year}/${day}` : "scratch";
  return `aoc-runner:code:${puzzle}:${language}`;
}

export function App() {
  const [year, setYear] = useState<string | null>(null);
  const [day, setDay] = useState<number | null>(null);
  const [refresh, setRefresh] = useState(0);

  const [languages, setLanguages] = useState<Language[]>([]);
  const [language, setLanguage] = useState(() => load("aoc-runner:language") ?? "python");
  const [code, setCode] = useState("");

  const [running, setRunning] = useState(false);
  const [result, setResult] = useState<Result | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    getLanguages()
      .then(setLanguages)
      .catch((err: Error) => setError(err.message));
  }, []);

  // load the saved code, or the starting code, whenever the puzzle or language changes
  useEffect(() => {
    const defaultCode = languages.find((l) => l.name === language)?.default_code ?? "";
    setCode(load(codeKey(year, day, language)) ?? defaultCode);
    setResult(null);
    setError(null);
  }, [year, day, language, languages]);

  function selectPuzzle(year: string, day: number | null) {
    setYear(year);
    setDay(day);
  }

  function changeLanguage(language: string) {
    setLanguage(language);
    save("aoc-runner:language", language);
  }

  function changeCode(code: string) {
    setCode(code);
    save(codeKey(year, day, language), code);
  }

  async function run() {
    if (running || !code.trim()) {
      return;
    }

    setRunning(true);
    setError(null);

    try {
      setResult(await submit(language, code));
    } catch (err) {
      setResult(null);
      setError((err as Error).message);
    } finally {
      setRunning(false);
    }
  }

  return (
    <div className="app">
      <header className="topbar">
        <div className="brand">
          <span className="brand-mark" aria-hidden="true">
            ▸
          </span>
          aoc-runner
        </div>
        <AuthStatus onLogin={() => setRefresh((n) => n + 1)} />
      </header>

      <aside className="sidebar">
        <PuzzleBrowser year={year} day={day} onSelect={selectPuzzle} />
      </aside>

      <main className="content">
        <section className="panel puzzle" aria-label="Puzzle">
          <h2 className="panel-title">{year && day ? `${year} · Day ${day}` : "Puzzle"}</h2>
          {year && day ? (
            <ProblemView year={year} day={day} refresh={refresh} />
          ) : (
            <p className="muted">Pick a day to read the puzzle.</p>
          )}
        </section>

        <section className="panel workspace" aria-label="Solution">
          <Editor
            languages={languages}
            language={language}
            code={code}
            running={running}
            onLanguageChange={changeLanguage}
            onCodeChange={changeCode}
            onRun={run}
          />
          <ResultView result={result} error={error} />
        </section>
      </main>
    </div>
  );
}
