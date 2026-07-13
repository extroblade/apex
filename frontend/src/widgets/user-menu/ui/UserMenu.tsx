import { Link } from 'wouter';
import {
  Settings2,
  User,
  LogOut,
  LogIn,
  Palette,
  Languages,
  Sun,
  Moon,
  type LucideIcon,
} from 'lucide-react';

import { useViewer } from '@/entities/viewer';
import { useLogout } from '@/features/auth';
import { useThemeStore, THEMES, type Theme } from '@/shared/theme';
import { useTranslation, setLanguage, useAvailableLocales } from '@/shared/i18n';
import { Avatar, AvatarFallback, AvatarImage } from '@/shared/ui/avatar';
import { initials } from '@/shared/lib/initials';
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
} from '@/shared/ui/dropdown-menu';

/**
 * The account/preferences menu. Opens from the avatar when signed in (Profile,
 * Theme, Language, Log out), or from a settings icon when signed out (Theme,
 * Language, Log in). Theme + language switching live here.
 */
const THEME_ICONS: Record<Theme, LucideIcon> = {
  light: Sun,
  dark: Moon,
  custom: Palette,
};

export function UserMenu() {
  const { data: viewer } = useViewer();
  const logout = useLogout();
  const theme = useThemeStore((s) => s.theme);
  const setTheme = useThemeStore((s) => s.setTheme);
  const { t, i18n } = useTranslation();
  const locales = useAvailableLocales();

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        {viewer ? (
          <button
            className="cursor-pointer rounded-full outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background"
            aria-label={t('common.userMenu')}
          >
            <Avatar>
              {viewer.avatarUrl && <AvatarImage src={viewer.avatarUrl} alt="" />}
              <AvatarFallback>{initials(viewer.nickname || viewer.email)}</AvatarFallback>
            </Avatar>
          </button>
        ) : (
          <button
            className="inline-flex size-9 cursor-pointer items-center justify-center rounded-md border outline-none hover:bg-accent focus-visible:ring-2 focus-visible:ring-ring"
            aria-label={t('common.preferences')}
          >
            <Settings2 className="size-4" />
          </button>
        )}
      </DropdownMenuTrigger>

      <DropdownMenuContent align="end" className="w-60">
        {viewer && (
          <>
            <DropdownMenuLabel className="flex items-center gap-2">
              <Avatar className="size-7">
                {viewer.avatarUrl && <AvatarImage src={viewer.avatarUrl} alt="" />}
                <AvatarFallback>
                  {initials(viewer.nickname || viewer.email)}
                </AvatarFallback>
              </Avatar>
              <span className="min-w-0 truncate">{viewer.nickname || viewer.email}</span>
            </DropdownMenuLabel>
            <DropdownMenuSeparator />
            <DropdownMenuItem asChild>
              <Link href="/profile">
                <User className="size-4 text-muted-foreground" />
                {t('userMenu.profile')}
              </Link>
            </DropdownMenuItem>
            <DropdownMenuSeparator />
          </>
        )}

        <DropdownMenuLabel className="flex items-center gap-2 text-xs text-muted-foreground">
          <Palette className="size-3.5" />
          {t('userMenu.theme')}
        </DropdownMenuLabel>
        <DropdownMenuRadioGroup value={theme} onValueChange={(v) => setTheme(v as Theme)}>
          {THEMES.map((tm) => {
            const Icon = THEME_ICONS[tm];
            return (
              <DropdownMenuRadioItem key={tm} value={tm}>
                <span className="flex items-center gap-2">
                  <Icon className="size-3.5" />
                  {t(`theme.${tm}`)}
                </span>
              </DropdownMenuRadioItem>
            );
          })}
        </DropdownMenuRadioGroup>

        <DropdownMenuSeparator />
        <DropdownMenuLabel className="flex items-center gap-2 text-xs text-muted-foreground">
          <Languages className="size-3.5" />
          {t('userMenu.language')}
        </DropdownMenuLabel>
        <DropdownMenuRadioGroup
          value={i18n.language}
          onValueChange={(v) => void setLanguage(v)}
        >
          {locales.map((lng) => (
            <DropdownMenuRadioItem key={lng.code} value={lng.code}>
              {lng.name}
            </DropdownMenuRadioItem>
          ))}
        </DropdownMenuRadioGroup>

        <DropdownMenuSeparator />
        {viewer ? (
          <DropdownMenuItem onClick={() => logout.mutate()}>
            <LogOut className="size-4 text-muted-foreground" />
            {t('common.logOut')}
          </DropdownMenuItem>
        ) : (
          <DropdownMenuItem asChild>
            <Link href="/login">
              <LogIn className="size-4 text-muted-foreground" />
              {t('common.logIn')}
            </Link>
          </DropdownMenuItem>
        )}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
