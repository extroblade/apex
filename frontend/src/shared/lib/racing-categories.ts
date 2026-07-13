import { Car, Gauge, Flag, Mountain, Trees, CircleDot, type LucideIcon } from 'lucide-react';

export interface CategoryMeta {
  label: string;
  Icon: LucideIcon;
  hue: number;
}

const CATEGORY_META: Record<string, CategoryMeta> = {
  sports_car: { label: 'Sports Car', Icon: Car, hue: 12 },
  formula_car: { label: 'Formula', Icon: Gauge, hue: 210 },
  oval: { label: 'Oval', Icon: CircleDot, hue: 150 },
  road: { label: 'Road', Icon: Flag, hue: 190 },
  dirt_oval: { label: 'Dirt Oval', Icon: Mountain, hue: 32 },
  dirt_road: { label: 'Dirt Road', Icon: Trees, hue: 96 },
};

export function categoryMeta(category: string): CategoryMeta {
  return CATEGORY_META[category] ?? { label: category || 'Other', Icon: Flag, hue: 265 };
}
