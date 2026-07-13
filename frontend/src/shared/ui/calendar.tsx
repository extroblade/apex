import { DayPicker, getDefaultClassNames, type DayPickerProps } from 'react-day-picker';
import { ru as ruLocale } from 'react-day-picker/locale';

import { useTranslation } from '@/shared/i18n';
import { cn } from '@/shared/lib/utils';

/**
 * Token-themed react-day-picker. The locale follows the app language so
 * weekday/month names switch together with the rest of the UI.
 */
function Calendar({ className, classNames, ...props }: DayPickerProps) {
  const { i18n } = useTranslation();
  const defaults = getDefaultClassNames();

  return (
    <DayPicker
      locale={i18n.language.startsWith('ru') ? ruLocale : undefined}
      showOutsideDays
      className={cn('select-none', className)}
      classNames={{
        ...defaults,
        root: cn(defaults.root, 'text-sm'),
        months: cn(defaults.months, 'relative'),
        month_caption: cn(defaults.month_caption, 'flex h-8 items-center px-2 font-medium'),
        nav: cn(defaults.nav, 'absolute right-0 top-0 flex gap-1'),
        button_previous: cn(
          defaults.button_previous,
          'inline-flex size-7 cursor-pointer items-center justify-center rounded-md hover:bg-accent',
        ),
        button_next: cn(
          defaults.button_next,
          'inline-flex size-7 cursor-pointer items-center justify-center rounded-md hover:bg-accent',
        ),
        chevron: 'size-4 fill-current text-muted-foreground',
        weekday: 'w-9 pb-1 text-xs font-normal text-muted-foreground',
        day: 'p-0 text-center',
        day_button:
          'inline-flex size-9 cursor-pointer items-center justify-center rounded-md hover:bg-accent focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring',
        selected: 'rounded-md bg-primary text-primary-foreground hover:[&>button]:bg-primary',
        today: 'font-semibold text-primary',
        outside: 'text-muted-foreground/50',
        disabled: 'text-muted-foreground/40',
        ...classNames,
      }}
      {...props}
    />
  );
}

export { Calendar };
