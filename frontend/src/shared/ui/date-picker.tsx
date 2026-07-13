import { useState } from 'react';
import { CalendarIcon, X } from 'lucide-react';

import { useTranslation } from '@/shared/i18n';
import { cn } from '@/shared/lib/utils';
import { Calendar } from './calendar';
import { Popover, PopoverContent, PopoverTrigger } from './popover';

/** YYYY-MM-DD in local time (toISOString would shift across midnight UTC). */
function toISODate(d: Date): string {
  const p = (n: number) => String(n).padStart(2, '0');
  return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())}`;
}

function parseISODate(s: string): Date | undefined {
  const m = /^(\d{4})-(\d{2})-(\d{2})$/.exec(s);
  if (!m) return undefined;
  return new Date(Number(m[1]), Number(m[2]) - 1, Number(m[3]));
}

/**
 * A button-triggered calendar picker holding an ISO date string (''
 * = unset), designed to drop into react-hook-form via Controller.
 */
export function DatePicker({
  value,
  onChange,
  id,
  placeholder,
  clearable = true,
  className,
}: {
  value: string;
  onChange: (v: string) => void;
  id?: string;
  placeholder?: string;
  clearable?: boolean;
  className?: string;
}) {
  const { t, i18n } = useTranslation();
  const [open, setOpen] = useState(false);
  const selected = value ? parseISODate(value) : undefined;

  const label = selected
    ? selected.toLocaleDateString(i18n.language, {
        day: 'numeric',
        month: 'short',
        year: 'numeric',
      })
    : (placeholder ?? t('common.pickDate'));

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <div className={cn('flex items-center gap-1', className)}>
        <PopoverTrigger
          id={id}
          className={cn(
            'flex h-9 w-full cursor-pointer items-center gap-2 rounded-md border border-input bg-background px-3 py-2 text-sm outline-none',
            'focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background',
            !selected && 'text-muted-foreground',
          )}
        >
          <CalendarIcon className="size-4 shrink-0 opacity-60" />
          {label}
        </PopoverTrigger>
        {clearable && selected && (
          <button
            type="button"
            onClick={() => onChange('')}
            aria-label={t('common.clear')}
            className="inline-flex size-8 shrink-0 cursor-pointer items-center justify-center rounded-md text-muted-foreground hover:bg-accent hover:text-accent-foreground"
          >
            <X className="size-4" />
          </button>
        )}
      </div>
      <PopoverContent className="w-auto">
        <Calendar
          mode="single"
          selected={selected}
          defaultMonth={selected}
          onSelect={(d) => {
            onChange(d ? toISODate(d) : '');
            setOpen(false);
          }}
        />
      </PopoverContent>
    </Popover>
  );
}
