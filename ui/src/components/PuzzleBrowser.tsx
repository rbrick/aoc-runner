import { useEffect, useRef, useState } from "react";
import { getDays, getYears } from "../api";

type Props = {
  year: string | null;
  day: number | null;
  onSelect: (year: string, day: number | null) => void;
};

export function PuzzleBrowser({ year, day, onSelect }: Props) {
  const [years, setYears] = useState<string[]>([]);
  const [days, setDays] = useState<number[] | null>(null);
  const [error, setError] = useState<string | null>(null);

  // the newest year can be listed before any of its puzzles are out,
  // so on first load keep going back until a year has puzzles
  const firstLoad = useRef(true);

  useEffect(() => {
    getYears()
      .then((years) => {
        setYears(years);

        if (!year && years.length > 0) {
          onSelect(years[0], null);
        }
      })
      .catch((err: Error) => setError(err.message));
  }, []);

  useEffect(() => {
    if (!year) {
      return;
    }

    setDays(null);

    getDays(year)
      .then((days) => {
        const next = years[years.indexOf(year) + 1];

        if (firstLoad.current && days.length === 0 && next) {
          onSelect(next, null);
          return;
        }

        firstLoad.current = false;
        setDays(days);
      })
      .catch((err: Error) => setError(err.message));
  }, [year, years]);

  return (
    <nav className="browser" aria-label="Puzzles">
      <label className="field">
        <span className="field-label">Year</span>
        <select value={year ?? ""} onChange={(event) => onSelect(event.target.value, null)}>
          {years.map((year) => (
            <option key={year} value={year}>
              {year}
            </option>
          ))}
        </select>
      </label>

      {error && <p className="muted">{error}</p>}

      {days?.length === 0 && <p className="muted">No puzzles have been released for {year} yet.</p>}

      <ol className="days">
        {(days ?? []).map((number) => (
          <li key={number}>
            <button
              className={number === day ? "day day-selected" : "day"}
              aria-current={number === day ? "page" : undefined}
              onClick={() => year && onSelect(year, number)}
            >
              {number}
            </button>
          </li>
        ))}
      </ol>
    </nav>
  );
}
