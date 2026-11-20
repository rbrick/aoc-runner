import { useState } from "react";
import { AuthStatus } from "./components/AuthStatus";
import { PuzzleView } from "./components/PuzzleView";
import { YearsView } from "./components/YearsView";
import { YearView } from "./components/YearView";
import { href, useRoute } from "./router";

export function App() {
  const route = useRoute();

  // bumped after logging in or out, so stars, part two and inputs are fetched again
  const [refresh, setRefresh] = useState(0);

  return (
    <div className={route.view === "puzzle" ? "app app-workspace" : "app"}>
      <header className="topbar">
        <nav className="breadcrumbs" aria-label="Breadcrumb">
          <a className="brand" href={href()}>
            <span className="brand-mark" aria-hidden="true">
              ▸
            </span>
            aoc-runner
          </a>
          {route.view !== "years" && (
            <>
              <span className="breadcrumb-separator" aria-hidden="true">
                /
              </span>
              <a href={href(route.year)} aria-current={route.view === "year" ? "page" : undefined}>
                {route.year}
              </a>
            </>
          )}
          {route.view === "puzzle" && (
            <>
              <span className="breadcrumb-separator" aria-hidden="true">
                /
              </span>
              <span aria-current="page">Day {route.day}</span>
            </>
          )}
        </nav>
        <AuthStatus onLogin={() => setRefresh((n) => n + 1)} onLogout={() => setRefresh((n) => n + 1)} />
      </header>

      <main className="main">
        {route.view === "years" && <YearsView refresh={refresh} />}
        {route.view === "year" && <YearView year={route.year} refresh={refresh} />}
        {route.view === "puzzle" && <PuzzleView year={route.year} day={route.day} refresh={refresh} />}
      </main>
    </div>
  );
}
