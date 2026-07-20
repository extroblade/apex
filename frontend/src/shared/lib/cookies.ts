const DAY = 86400;

export function getCookie(name: string): string | undefined {
  const m = document.cookie.match(new RegExp('(?:^|; )' + name + '=([^;]*)'));
  return m ? decodeURIComponent(m[1]) : undefined;
}

export function setCookie(name: string, value: string, maxAgeSec = 365 * DAY): void {
  // Mark the cookie Secure on HTTPS origins so it never leaks over plaintext.
  // The session cookie (set by the Go API) already does this via COOKIE_SECURE;
  // this covers the developer cookie and any other client-set cookies.
  const secure = window.isSecureContext ? ';secure' : '';
  document.cookie = `${name}=${encodeURIComponent(value)};path=/;max-age=${maxAgeSec};samesite=lax${secure}`;
}

export function deleteCookie(name: string): void {
  const secure = window.isSecureContext ? ';secure' : '';
  document.cookie = `${name}=;path=/;max-age=0${secure}`;
}
