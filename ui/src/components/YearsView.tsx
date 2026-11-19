import { useEffect, useState } from "react";
import { getYears, type Year } from "../api";
import { href } from "../router";

type Props = {
  refresh: number;
};

export function YearsView({ refresh }: Props) {
  const [years, setYears] = useState<Year[] | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    getYears()
      .then(setYears)
      .catch((err: Error) => setError(err.message));
  }, [refresh]);

  return (
    <div className="page">
      <h1 className="page-title">Events</h1>

      {error && <p className="muted">Couldn't load the events: {error}</p>}
      {!years && !error && <p className="muted">Loading events…</p>}

      <ul className="year-grid">
        {years?.map((year) => (
          <li key={year.year}>
            <a className="card year-card" href={href(year.year)}>
              <span className="year-card-year">{year.year}</span>
              {year.stars !== undefined && (
                <span className="year-card-stars">
                  <span className="star star-earned" aria-hidden="true">
                    ★
                  </span>{" "}
                  {year.stars}
                </span>
              )}
            </a>
          </li>
        ))}
      </ul>
    </div>
  );
}
