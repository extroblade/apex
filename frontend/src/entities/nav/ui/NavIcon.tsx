import {
  Home,
  Fuel,
  CalendarRange,
  CalendarCheck,
  Warehouse,
  Wrench,
  Target,
  Users,
  Gauge,
  BarChart3,
  Circle,
  type LucideIcon,
} from 'lucide-react';

/**
 * Whitelist mapping the backend's icon *name* to a lucide component — the menu
 * is data, so it can only name an icon, never ship one. An unknown name (e.g. a
 * new item added from the Cockpit) degrades to a neutral dot instead of
 * breaking the menu.
 */
const ICONS: Record<string, LucideIcon> = {
  home: Home,
  fuel: Fuel,
  'calendar-range': CalendarRange,
  'calendar-check': CalendarCheck,
  warehouse: Warehouse,
  wrench: Wrench,
  target: Target,
  users: Users,
  gauge: Gauge,
  'bar-chart-3': BarChart3,
};

export function NavIcon({ name, className }: { name: string; className?: string }) {
  const Icon = ICONS[name] ?? Circle;
  return <Icon className={className} aria-hidden />;
}
