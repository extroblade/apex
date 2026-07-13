// Mirrors the Go structs in backend/internal/fuel.
export type RaceType = 'laps' | 'time';
export type FuelStrategy = 'balanced' | 'undercut' | 'overcut';
export type WindowUnit = 'laps' | 'minutes';

export interface PitWindow {
  from: number; // 0 = no lower bound
  to: number; // 0 = no upper bound
  unit: WindowUnit;
}

export interface FuelRules {
  mandatoryStops: number;
  windows: PitWindow[];
}

export interface FuelRequest {
  raceType: RaceType;
  strategy: FuelStrategy;
  raceLaps: number;
  raceMinutes: number;
  lapTimeSec: number;
  fuelPerLap: number;
  tankCapacity: number;
  extraLaps: number;
  rules: FuelRules;
}

export interface Stint {
  index: number;
  laps: number;
  fuel: number;
  pitOnLap: number; // 0 = race end
}

export interface FuelPlan {
  totalLaps: number;
  lapsPerStint: number;
  fuelNeeded: number;
  fuelWithMargin: number;
  startFuel: number;
  pitStops: number;
  stints: Stint[];
}
