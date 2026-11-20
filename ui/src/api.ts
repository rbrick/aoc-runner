export type Session = {
  logged_in: boolean;
  login_url?: string;
  message?: string;
};

export type Problem = {
  body: string;
  examples: string[] | null;
  example_answers: string[] | null;
};

export type Language = {
  name: string;
  default_code: string;
};

export type Result = {
  // false when the code failed to compile or failed while running
  ok: boolean;
  build: { duration_ms: number; errors: string[] | null };
  run?: {
    duration_ms: number;
    part_one: string;
    part_two: string;
    errors: string[] | null;
  };
};

export class ApiError extends Error {}

async function request<T>(path: string, init?: RequestInit, accept: number[] = []): Promise<T> {
  const response = await fetch(path, init);
  const body = await response.json().catch(() => null);

  if (!response.ok && !accept.includes(response.status)) {
    throw new ApiError(body?.message ?? `${response.status} ${response.statusText}`);
  }

  return body as T;
}

// stars are only included when logged in
export type Year = {
  year: string;
  stars?: number;
};

export type Day = {
  number: number;
  stars?: number;
};

export async function getYears(): Promise<Year[]> {
  const body = await request<{ events: Year[] | null }>("/events/");
  return (body.events ?? []).sort((a, b) => b.year.localeCompare(a.year));
}

export async function getDays(year: string): Promise<Day[]> {
  const body = await request<{ days: Day[] | null }>(`/events/${year}`);
  return body.days ?? [];
}

export class LoginRequiredError extends ApiError {}

// the user's own puzzle input, the backend downloads it once with the AoC session
export async function getInput(year: string, day: number): Promise<string> {
  const response = await fetch(`/events/${year}/${day}/input`);

  if (response.status === 401) {
    throw new LoginRequiredError("Log in to Advent of Code to load your input.");
  }

  if (!response.ok) {
    const body = await response.json().catch(() => null);
    throw new ApiError(body?.message ?? `${response.status} ${response.statusText}`);
  }

  return response.text();
}

export function getProblem(year: string, day: number): Promise<Problem> {
  return request<Problem>(`/events/${year}/${day}`);
}

export async function getLanguages(): Promise<Language[]> {
  const body = await request<{ languages: Language[] }>("/submit/languages");
  return body.languages;
}

// 422 means the code failed to build or run, the body still has the output
export async function submit(language: string, code: string, input: string): Promise<Result> {
  const response = await fetch("/submit/", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ language, code, input }),
  });

  const body = await response.json().catch(() => null);

  if (!response.ok && response.status !== 422) {
    throw new ApiError(body?.message ?? `${response.status} ${response.statusText}`);
  }

  return { ...body, ok: response.ok };
}

export function getSession(): Promise<Session> {
  return request<Session>("/auth/");
}

// 401 means no browser session was found, the body says where to log in
export function login(): Promise<Session> {
  return request<Session>("/auth/login", { method: "POST" }, [401]);
}

// only forgets the session on the server, adventofcode.com stays logged in
export function logout(): Promise<Session> {
  return request<Session>("/auth/logout", { method: "POST" });
}

export type SubmitResult = {
  outcome: "correct" | "incorrect" | "rate_limited" | "wrong_level" | "unknown";
  // AoC's own message
  message: string;
  wait_seconds?: number;
  // the answer was already rejected before, so it wasn't sent again
  duplicate?: boolean;
};

// level 1 is part one, level 2 is part two
export function submitAnswer(year: string, day: number, level: 1 | 2, answer: string): Promise<SubmitResult> {
  return request<SubmitResult>(`/events/${year}/${day}/answer`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ level, answer }),
  });
}
