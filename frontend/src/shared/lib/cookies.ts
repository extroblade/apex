const DAY = 86400;

export function getCookie(name: string): string | undefined {
  const m = document.cookie.match(new RegExp('(?:^|; )' + name + '=([^;]*)'));
  return m ? decodeURIComponent(m[1]) : undefined;
}

export function setCookie(name: string, value: string, maxAgeSec = 365 * DAY): void {
  document.cookie = `${name}=${encodeURIComponent(value)};path=/;max-age=${maxAgeSec};samesite=lax`;
}

export function deleteCookie(name: string): void {
  document.cookie = `${name}=;path=/;max-age=0`;
}
