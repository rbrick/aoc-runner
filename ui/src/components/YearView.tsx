import { useEffect, useState } from "react";
import { getDays, type Day } from "../api";
import { href } from "../router";
import { Stars } from "./Stars";

type Props = {
  year: string;
  refresh: number;
};

export function YearView({ year, refresh }: Props) {
  const [days, setDays] = useState<Day[] | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    setDays(null);
    setError(null);

    getDays(year)
      .then(setDays)
      .catch((err: Error) => setError(err.message));
  }, [year, refresh]);

  const showStars = days?.some((day) => day.stars !== undefined) ?? false;
  const earned = days?.reduce((total, day) => total + (day.stars ?? 0), 0) ?? 0;

  return (
    <div className="page">
      <div className="page-header">
        <h1 className="page-title">{year}</h1>
        {showStars && days && (
          <span className="muted">
            {earned} / {days.length * 2} stars
          </span>
        )}
      </div>

      {error && <p className="muted">Couldn't load the puzzles: {error}</p>}
      {!days && !error && <p className="muted">Loading puzzles…</p>}
      {days?.length === 0 && <p className="muted">No puzzles have been released for {year} yet.</p>}

      <ol className="day-grid">
        {days?.map((day) => (
          <li key={day.number}>
            <a
              className={day.stars === 2 ? "card day-card day-card-complete" : "card day-card"}
              href={href(year, day.number)}
            >
              <span className="day-card-label">Day</span>
              <span className="day-card-number">{day.number}</span>
              {day.stars !== undefined && <Stars earned={day.stars} />}
            </a>
          </li>
        ))}
      </ol>
    </div>
  );
}
