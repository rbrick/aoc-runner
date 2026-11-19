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

export async function getYears(): Promise<string[]> {
  const body = await request<{ events: { year: string }[] | null }>("/events/");
  return (body.events ?? []).map((event) => event.year).sort().reverse();
}

export async function getDays(year: string): Promise<number[]> {
  const body = await request<{ days: { day: string }[] | null }>(`/events/${year}`);

  return (body.days ?? [])
    .map((day) => Number.parseInt(day.day.replace(/\D/g, ""), 10))
    .filter((day) => !Number.isNaN(day))
    .sort((a, b) => a - b);
}

export function getProblem(year: string, day: number): Promise<Problem> {
  return request<Problem>(`/events/${year}/${day}`);
}

export async function getLanguages(): Promise<Language[]> {
  const body = await request<{ languages: Language[] }>("/submit/languages");
  return body.languages;
}

// 422 means the code failed to build or run, the body still has the output
export async function submit(language: string, code: string): Promise<Result> {
  const response = await fetch("/submit/", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ language, code }),
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
