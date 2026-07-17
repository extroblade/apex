/**
 * Known product events. Referencing these constants at call sites keeps event
 * names consistent and gives autocomplete + a single catalogue — but the counter
 * still accepts any string, so ad-hoc events don't need a registry entry first.
 */
export const MetricEvents = {
  baselineGenerated: 'setup_baseline_generated',
  packGenerated: 'setup_pack_generated',
  setupSaved: 'setup_saved',
  packSavedAll: 'setup_pack_saved_all',
  languageChanged: 'language_changed',
  racePlanned: 'race_planned',
  goalCreated: 'goal_created',
} as const;

export type KnownEvent = (typeof MetricEvents)[keyof typeof MetricEvents];

// Known events light up autocomplete; `string & {}` still permits any string
// without widening the type back to a plain `string` (which would kill it).
export type MetricEvent = KnownEvent | (string & {});
