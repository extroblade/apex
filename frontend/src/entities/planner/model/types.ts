export interface CarItem {
  carId: number;
  carName: string;
  category: string;
  description: string;
  owned: boolean;
  free: boolean;
}

export interface TrackItem {
  trackId: number;
  trackName: string;
  configName: string;
  category: string;
  description: string;
  owned: boolean;
  free: boolean;
}

/**
 * How a user can run a track:
 * - `free` — default/included content (green)
 * - `owned` — purchased or unlocked via a bundle (aquamarine)
 * - `missing` — not available (red)
 */
export type TrackAccess = 'free' | 'owned' | 'missing';

export interface SeriesItem {
  seriesId: number;
  seriesName: string;
  category: string;
  description: string;
  licenseNeeded: string;
  favorite: boolean;
}

export interface SeasonWeek {
  week: number;
  trackId: number;
  trackName: string;
  configName: string;
  /** Week-start date (YYYY-MM-DD) from the real schedule; may be absent. */
  raceDate?: string;
  /** True when the track is available at all (free or owned). */
  trackOwned: boolean;
  /** Finer-grained availability, drives the cell color. */
  trackAccess: TrackAccess;
  planned: boolean;
}

export interface SeasonSeries {
  seriesId: number;
  seriesName: string;
  category: string;
  licenseNeeded: string;
  favorite: boolean;
  carOwned: boolean;
  weeks: SeasonWeek[];
}

export interface Season {
  currentWeek: number;
  totalWeeks: number;
  series: SeasonSeries[];
}

export interface CatalogCounts {
  cars: number;
  tracks: number;
  series: number;
}
