/** Derives up-to-2-letter initials from a nickname or email for avatar fallbacks. */
export function initials(nameOrEmail: string): string {
  const base = nameOrEmail.trim();
  if (!base) return '?';
  const namePart = base.includes('@') ? base.split('@')[0] : base;
  const words = namePart.split(/[\s._-]+/).filter(Boolean);
  if (words.length >= 2) {
    return (words[0][0] + words[1][0]).toUpperCase();
  }
  return namePart.slice(0, 2).toUpperCase();
}
