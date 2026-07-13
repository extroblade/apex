export interface Goal {
  id: number;
  title: string;
  notes: string;
  unit: string;
  target: number;
  current: number;
  done: boolean;
  dueDate: string | null; // YYYY-MM-DD
  progress: number; // 0..1
  createdAt: string;
}

export interface GoalInput {
  title: string;
  notes: string;
  unit: string;
  target: number;
  current: number;
  done?: boolean;
  dueDate?: string | null;
}
