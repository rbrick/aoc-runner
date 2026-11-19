// browser storage can be unavailable (private windows, blocked site data), so never let it throw

export function load(key: string): string | null {
  try {
    return window.localStorage.getItem(key);
  } catch {
    return null;
  }
}

export function save(key: string, value: string) {
  try {
    window.localStorage.setItem(key, value);
  } catch {
    // nothing to do, it just won't be remembered
  }
}
