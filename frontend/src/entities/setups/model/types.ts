export interface Setup {
  id: number;
  carId: number;
  carName: string;
  trackId: number;
  trackName: string;
  name: string;
  notes: string;
  data?: string; // only present on single-setup fetches
  category: string;
  author: string;
  public: boolean;
  downloads: number;
  mine: boolean;
  createdAt: string;
}

export interface NewSetup {
  carId: number;
  trackId: number;
  name: string;
  notes: string;
  data: string;
  public: boolean;
}
