// Types mirroring the Go JSON. Note: the nested iRacing payloads (License,
// CareerStat, RecentRace) use snake_case because they pass through from the
// iRacing API unchanged; our own wrappers (Account, GroupStat) use camelCase.

export interface IRacingAccount {
  custId: number;
  displayName: string;
  linkedAt: string;
}

export interface IRacingStatus {
  linked: boolean;
  account?: IRacingAccount;
}

export interface License {
  category_id: number;
  category: string;
  license_level: number;
  safety_rating: number;
  irating: number;
  group_name: string;
}

export interface CareerStat {
  category: string;
  category_id: number;
  starts: number;
  wins: number;
  top5: number;
  poles: number;
  avg_start_position: number;
  avg_finish_position: number;
  laps: number;
  laps_led: number;
  avg_incidents: number;
  win_percentage: number;
}

export interface RecentRace {
  subsession_id: number;
  series_id: number;
  series_name: string;
  session_start_time: string;
  car_id: number;
  track: { track_id: number; track_name: string };
  start_position: number;
  finish_position: number;
  incidents: number;
  oldi_rating: number;
  newi_rating: number;
  laps_complete: number;
  category_id: number;
}

export interface Dashboard {
  account: IRacingAccount;
  licenses: License[];
  career: CareerStat[];
  recent: RecentRace[];
}

export type CompareDimension = 'categories' | 'cars' | 'tracks';

export interface GroupStat {
  key: number;
  label: string;
  races: number;
  avgFinish: number;
  avgStart: number;
  avgIncidents: number;
  avgIRatingGain: number;
}
