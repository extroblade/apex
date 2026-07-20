import { useEffect, useMemo, useRef, useState } from 'react';

import { Input } from '@/shared/ui/input';
import { Card, CardContent } from '@/shared/ui/card';
import { SkeletonRows } from '@/shared/ui/skeleton';
import { CatalogThumbnail } from '@/shared/ui/catalog-thumbnail';
import { Badge } from '@/shared/ui/badge';
import { categoryMeta } from '@/shared/lib/racing-categories';
import { cn } from '@/shared/lib/utils';
import { CatalogInfoButton, type TrackLayout } from '@/entities/planner';

const PAGE = 20;

export interface CatalogRow {
  id: number;
  name: string;
  category: string;
  description: string;
  sub?: string;
  badge?: string; // e.g. required license class for a series
  layouts?: TrackLayout[]; // for tracks: the configs grouped under this row
  checked: boolean;
  // `disabled` freezes the row (checked + read-only) — used for free content,
  // which everyone has and can't be "unowned".
  disabled?: boolean;
  // `free` flags the row for a "Free" badge so the read-only state is clear.
  free?: boolean;
}

export function CatalogList({
  items,
  loading,
  onToggle,
}: {
  items: CatalogRow[];
  loading: boolean;
  onToggle: (id: number, checked: boolean) => void;
}) {
  const [query, setQuery] = useState('');
  const [category, setCategory] = useState<string | null>(null);
  const [limit, setLimit] = useState(PAGE);

  const categories = useMemo(
    () => Array.from(new Set(items.map((i) => i.category).filter(Boolean))).sort(),
    [items],
  );

  const filtered = useMemo(
    () =>
      items.filter(
        (i) =>
          (!category || i.category === category) &&
          `${i.name} ${i.sub ?? ''}`.toLowerCase().includes(query.toLowerCase()),
      ),
    [items, category, query],
  );

  // Reset the visible window when the filter/search changes.
  useEffect(() => setLimit(PAGE), [query, category]);

  // Infinite scroll: grow the window when the sentinel is reached.
  const sentinel = useRef<HTMLDivElement>(null);
  useEffect(() => {
    const el = sentinel.current;
    if (!el) return;
    const obs = new IntersectionObserver((entries) => {
      if (entries[0].isIntersecting) setLimit((l) => l + PAGE);
    });
    obs.observe(el);
    return () => obs.disconnect();
  }, [filtered.length]);

  if (loading) {
    return (
      <Card>
        <CardContent className="pt-6">
          <SkeletonRows rows={8} />
        </CardContent>
      </Card>
    );
  }
  if (!items.length) {
    return (
      <Card>
        <CardContent className="py-6 text-sm text-muted-foreground">
          Nothing in the catalog yet.
        </CardContent>
      </Card>
    );
  }

  const visible = filtered.slice(0, limit);

  return (
    <div className="space-y-3">
      <Input
        placeholder="Search…"
        value={query}
        onChange={(e) => setQuery(e.target.value)}
      />

      <div className="flex flex-wrap gap-2">
        <FilterChip active={category === null} onClick={() => setCategory(null)}>
          All
        </FilterChip>
        {categories.map((c) => {
          const { label, Icon } = categoryMeta(c);
          return (
            <FilterChip key={c} active={category === c} onClick={() => setCategory(c)}>
              <Icon className="size-3.5" />
              {label}
            </FilterChip>
          );
        })}
      </div>

      <Card>
        <CardContent className="max-h-[26rem] overflow-y-auto pt-4">
          {visible.map((row) => (
            <Row key={row.id} row={row} onToggle={onToggle} />
          ))}
          {filtered.length === 0 && (
            <p className="py-4 text-sm text-muted-foreground">No matches.</p>
          )}
          {visible.length < filtered.length && (
            <div ref={sentinel} className="pt-2">
              <SkeletonRows rows={2} />
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

function FilterChip({
  active,
  onClick,
  children,
}: {
  active: boolean;
  onClick: () => void;
  children: React.ReactNode;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={cn(
        'inline-flex cursor-pointer items-center gap-1.5 rounded-full border px-3 py-1 text-xs font-medium transition-colors',
        active
          ? 'border-transparent bg-primary text-primary-foreground'
          : 'text-muted-foreground hover:bg-accent hover:text-accent-foreground',
      )}
    >
      {children}
    </button>
  );
}

function Row({
  row,
  onToggle,
}: {
  row: CatalogRow;
  onToggle: (id: number, checked: boolean) => void;
}) {
  return (
    <div className="flex items-center gap-3 border-b py-2 last:border-0">
      <CatalogThumbnail category={row.category} name={row.name} className="size-9" />
      <label
        className={cn(
          'flex flex-1 items-center gap-3',
          row.disabled ? 'cursor-default' : 'cursor-pointer',
        )}
      >
        <input
          type="checkbox"
          className="size-4 cursor-pointer disabled:cursor-default disabled:opacity-60"
          checked={row.checked}
          disabled={row.disabled}
          onChange={(e) => onToggle(row.id, e.target.checked)}
        />
        <span className="text-sm">
          {row.name}
          {row.sub && <span className="text-muted-foreground"> · {row.sub}</span>}
        </span>
      </label>
      {row.free && (
        <Badge className="bg-emerald-500/15 text-emerald-600 dark:text-emerald-400">
          Free
        </Badge>
      )}
      {row.badge && <Badge>{row.badge}</Badge>}
      {row.layouts && row.layouts.length > 1 && (
        <Badge className="tabular-nums" title={`${row.layouts.length} layouts`}>
          {row.layouts.length}×
        </Badge>
      )}
      <CatalogInfoButton
        name={row.name}
        category={row.category}
        description={row.description}
        sub={row.badge ? [row.sub, row.badge].filter(Boolean).join(' · ') : row.sub}
        layouts={row.layouts}
      />
    </div>
  );
}
