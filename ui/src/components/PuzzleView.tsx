import { useEffect, useRef, useState } from "react";
import {
  getInput,
  getLanguages,
  getProblem,
  LoginRequiredError,
  submit,
  submitAnswer,
  type Language,
  type Problem,
  type Result,
} from "../api";
import { load, save } from "../storage";
import { Editor } from "./Editor";
import { InputPanel, type InputSource, type MyInput } from "./InputPanel";
import { ProblemView } from "./ProblemView";
import { ResultView } from "./ResultView";

// code is remembered per puzzle and language
function codeKey(year: string, day: number, language: string) {
  return `aoc-runner:code:${year}/${day}:${language}`;
}

function inputKey(year: string, day: number) {
  return `aoc-runner:input:${year}/${day}`;
}

type Props = {
  year: string;
  day: number;
  // bumped after logging in, so part two and your input become available
  refresh: number;
};

export function PuzzleView({ year, day, refresh }: Props) {
  const [problem, setProblem] = useState<Problem | null>(null);
  const [problemError, setProblemError] = useState<string | null>(null);

  const [languages, setLanguages] = useState<Language[]>([]);
  const [language, setLanguage] = useState(() => load("aoc-runner:language") ?? "python");
  const [code, setCode] = useState("");

  const [inputSource, setInputSource] = useState<InputSource>("custom");
  const [customInput, setCustomInput] = useState("");
  const [myInput, setMyInput] = useState<MyInput>({ status: "idle" });

  // bumped after a correct answer, so part two shows up
  const [problemVersion, setProblemVersion] = useState(0);
  const startedOnExample = useRef(false);

  const [running, setRunning] = useState(false);
  const [runCount, setRunCount] = useState(0);
  const [result, setResult] = useState<Result | null>(null);
  const [resultSource, setResultSource] = useState<InputSource>("custom");
  const [error, setError] = useState<string | null>(null);

  const examples = problem?.examples ?? [];

  let input = customInput;

  if (inputSource === "mine") {
    input = myInput.status === "ready" ? myInput.input : "";
  } else if (typeof inputSource === "number") {
    input = examples[inputSource] ?? "";
  }

  useEffect(() => {
    getLanguages()
      .then(setLanguages)
      .catch((err: Error) => setError(err.message));
  }, []);

  // a new puzzle starts with a clean slate
  useEffect(() => {
    setProblem(null);
    setProblemError(null);
    setCustomInput(load(inputKey(year, day)) ?? "");
    setInputSource("custom");
    setMyInput({ status: "idle" });
    startedOnExample.current = false;
  }, [year, day]);

  // reloading the puzzle, after logging in or solving a part, keeps the input that's picked
  useEffect(() => {
    let cancelled = false;

    getProblem(year, day)
      .then((problem) => {
        if (cancelled) {
          return;
        }

        setProblem(problem);
        setProblemError(null);

        // start on the first example, unless custom input was saved for this puzzle
        if (!startedOnExample.current) {
          startedOnExample.current = true;

          if ((problem.examples ?? []).length > 0 && !load(inputKey(year, day))) {
            setInputSource(0);
          }
        }
      })
      .catch((err: Error) => !cancelled && setProblemError(err.message));

    return () => {
      cancelled = true;
    };
  }, [year, day, refresh, problemVersion]);

  // your input is only downloaded when it's picked
  useEffect(() => {
    // a request still loading is restarted, its result would be dropped by the cleanup
    if (inputSource !== "mine" || myInput.status === "ready") {
      return;
    }

    let cancelled = false;

    setMyInput({ status: "loading" });

    getInput(year, day)
      .then((input) => !cancelled && setMyInput({ status: "ready", input }))
      .catch((err: Error) => {
        if (!cancelled) {
          const message = err instanceof LoginRequiredError ? err.message : `Couldn't load your input: ${err.message}`;
          setMyInput({ status: "error", message });
        }
      });

    return () => {
      cancelled = true;
    };
    // refresh retries after logging in
  }, [inputSource, year, day, refresh]);

  // load the saved code, or the starting code, whenever the puzzle or language changes
  useEffect(() => {
    const defaultCode = languages.find((l) => l.name === language)?.default_code ?? "";
    setCode(load(codeKey(year, day, language)) ?? defaultCode);
    setResult(null);
    setError(null);
  }, [year, day, language, languages]);

  function changeLanguage(language: string) {
    setLanguage(language);
    save("aoc-runner:language", language);
  }

  function changeCode(code: string) {
    setCode(code);
    save(codeKey(year, day, language), code);
  }

  function changeCustomInput(input: string) {
    setInputSource("custom");
    setCustomInput(input);
    save(inputKey(year, day), input);
  }

  async function run() {
    if (running || !code.trim()) {
      return;
    }

    setRunning(true);
    setError(null);

    try {
      setResult(await submit(language, code, input));
      setResultSource(inputSource);
      setRunCount((n) => n + 1);
    } catch (err) {
      setResult(null);
      setError((err as Error).message);
    } finally {
      setRunning(false);
    }
  }

  // the example answers only make sense to compare against when an example was the input
  const expected = typeof resultSource === "number" ? (problem?.example_answers ?? null) : null;

  // only answers for your own input can be submitted
  const submitPart =
    resultSource === "mine" && result?.ok
      ? (level: 1 | 2, answer: string) => submitAnswer(year, day, level, answer)
      : null;

  return (
    <div className="workspace-layout">
      <section className="panel puzzle" aria-label="Puzzle">
        <div className="panel-header">
          <h2 className="panel-title">
            {year} · Day {day}
          </h2>
          <a
            className="external-link"
            href={`https://adventofcode.com/${year}/day/${day}`}
            target="_blank"
            rel="noreferrer"
          >
            Open on adventofcode.com <span aria-hidden="true">↗</span>
          </a>
        </div>
        <ProblemView problem={problem} error={problemError} onUseExample={setInputSource} />
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
        <InputPanel
          examples={examples}
          source={inputSource}
          value={input}
          myInput={myInput}
          onSourceChange={setInputSource}
          onCustomChange={changeCustomInput}
        />
        <ResultView
          key={runCount}
          result={result}
          error={error}
          expected={expected}
          submitPart={submitPart}
          onSolved={() => setProblemVersion((n) => n + 1)}
        />
      </section>
    </div>
  );
}
