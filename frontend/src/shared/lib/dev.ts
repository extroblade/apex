import { getCookie } from './cookies';

const DEV_COOKIE = 'developer';

/** Returns the developer key stored in the cookie (set via ?dev=KEY), or undefined. */
export function isDev(): string | undefined {
  return getCookie(DEV_COOKIE);
}

/**
 * console.log wrapper that is a no-op unless the developer cookie is set.
 * This keeps debug logs out of production (no cookie = no output).
 */
export function devlog(...args: unknown[]): void {
  if (getCookie(DEV_COOKIE)) {
    // eslint-disable-next-line no-console
    console.log(...args);
  }
}
