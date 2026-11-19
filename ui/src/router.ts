import { useEffect, useState } from "react";

// hash routes, so the Go server only ever has to serve index.html
//   #/          years
//   #/2024      a year's puzzles
//   #/2024/5    a puzzle
export type Route =
  | { view: "years" }
  | { view: "year"; year: string }
  | { view: "puzzle"; year: string; day: number };

function parse(hash: string): Route {
  const [year, day] = hash.replace(/^#\/?/, "").split("/").filter(Boolean);

  if (year && /^\d{4}$/.test(year)) {
    const number = Number(day);

    if (day && Number.isInteger(number) && number > 0) {
      return { view: "puzzle", year, day: number };
    }

    return { view: "year", year };
  }

  return { view: "years" };
}

export function useRoute(): Route {
  const [route, setRoute] = useState(() => parse(window.location.hash));

  useEffect(() => {
    const onChange = () => {
      setRoute(parse(window.location.hash));
      window.scrollTo(0, 0);
    };

    window.addEventListener("hashchange", onChange);
    return () => window.removeEventListener("hashchange", onChange);
  }, []);

  return route;
}

export function href(year?: string, day?: number) {
  if (!year) {
    return "#/";
  }

  return day ? `#/${year}/${day}` : `#/${year}`;
}
