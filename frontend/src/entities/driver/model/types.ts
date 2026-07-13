import type { License, CareerStat, RecentRace } from '@/entities/iracing';

export interface DriverSearchResult {
  cust_id: number;
  display_name: string;
}

export interface DriverProfile {
  custId: number;
  displayName: string;
  licenses: License[];
  career: CareerStat[];
  recent: RecentRace[];
  cachedAt: string;
}
