import { useState } from 'react';

import { useTranslation } from '@/shared/i18n';
import {
  useThemeStore,
  readCustomVars,
  saveCustomVars,
  applyCustomVars,
  DEFAULT_CUSTOM_VARS,
  type CustomThemeVars,
  type FontKey,
} from '@/shared/theme';
import { Button } from '@/shared/ui/button';
import { Label } from '@/shared/ui/label';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/shared/ui/card';

const FONTS: FontKey[] = ['system', 'serif', 'mono', 'rounded'];

const FONT_LABEL: Record<FontKey, string> = {
  system: 'themeConfig.fontSystem',
  serif: 'themeConfig.fontSerif',
  mono: 'themeConfig.fontMono',
  rounded: 'themeConfig.fontRounded',
};

/** Color + font configurator for the `custom` theme; persists to localStorage. */
export function ThemeConfigurator() {
  const { t } = useTranslation();
  const theme = useThemeStore((s) => s.theme);
  const setTheme = useThemeStore((s) => s.setTheme);
  const [vars, setVars] = useState<CustomThemeVars>(readCustomVars);

  const update = (patch: Partial<CustomThemeVars>) => {
    const next = { ...vars, ...patch };
    setVars(next);
    saveCustomVars(next);
    applyCustomVars(theme); // live-preview when the custom theme is active
  };

  const colors: Array<{ key: keyof CustomThemeVars & string; label: string }> = [
    { key: 'background', label: t('themeConfig.background') },
    { key: 'foreground', label: t('themeConfig.foreground') },
    { key: 'primary', label: t('themeConfig.primary') },
    { key: 'accent', label: t('themeConfig.accent') },
  ];

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t('themeConfig.title')}</CardTitle>
        <CardDescription>{t('themeConfig.desc')}</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {theme !== 'custom' && (
          <Button variant="outline" size="sm" onClick={() => setTheme('custom')}>
            {t('themeConfig.activate')}
          </Button>
        )}

        <div className="grid grid-cols-2 gap-4 sm:grid-cols-4">
          {colors.map(({ key, label }) => (
            <div key={key} className="space-y-1.5">
              <Label htmlFor={`tc-${key}`}>{label}</Label>
              <input
                id={`tc-${key}`}
                type="color"
                value={String(vars[key])}
                onChange={(e) => update({ [key]: e.target.value })}
                className="h-9 w-full cursor-pointer rounded-md border bg-transparent p-1"
              />
            </div>
          ))}
        </div>

        <div className="space-y-1.5">
          <Label>{t('themeConfig.font')}</Label>
          <div className="flex flex-wrap gap-2" role="group" aria-label={t('themeConfig.font')}>
            {FONTS.map((f) => (
              <Button
                key={f}
                type="button"
                size="sm"
                variant={vars.font === f ? 'default' : 'outline'}
                aria-pressed={vars.font === f}
                onClick={() => update({ font: f })}
              >
                {t(FONT_LABEL[f])}
              </Button>
            ))}
          </div>
        </div>

        <Button variant="ghost" size="sm" onClick={() => update(DEFAULT_CUSTOM_VARS)}>
          {t('themeConfig.reset')}
        </Button>
      </CardContent>
    </Card>
  );
}
