import { useRef, useState } from 'react';

import { useViewer } from '@/entities/viewer';
import { useTranslation } from '@/shared/i18n';
import { Avatar, AvatarFallback, AvatarImage } from '@/shared/ui/avatar';
import { Button } from '@/shared/ui/button';
import { Input } from '@/shared/ui/input';
import { Label } from '@/shared/ui/label';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/shared/ui/card';
import { initials } from '@/shared/lib/initials';
import { useUpdateProfile, useUpdateAvatar, useChangePassword } from '../api/use-profile';

const MAX_AVATAR_BYTES = 512 * 1024;

export function ProfileSettings() {
  const { t } = useTranslation();
  const { data: viewer } = useViewer();

  if (!viewer) return null;

  return (
    <div className="space-y-6">
      <AvatarCard
        avatarUrl={viewer.avatarUrl}
        label={viewer.nickname || viewer.email}
      />
      <AccountCard nickname={viewer.nickname} email={viewer.email} />
      <PasswordCard />
      <span className="sr-only">{t('profile.title')}</span>
    </div>
  );
}

function AvatarCard({ avatarUrl, label }: { avatarUrl: string; label: string }) {
  const { t } = useTranslation();
  const avatar = useUpdateAvatar();
  const inputRef = useRef<HTMLInputElement>(null);
  const [error, setError] = useState('');

  const onFile = (e: React.ChangeEvent<HTMLInputElement>) => {
    setError('');
    const file = e.target.files?.[0];
    if (!file) return;
    if (file.size > MAX_AVATAR_BYTES) {
      setError(t('profile.imageTooLarge'));
      return;
    }
    const reader = new FileReader();
    reader.onload = () => avatar.mutate(String(reader.result));
    reader.readAsDataURL(file);
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t('profile.avatar')}</CardTitle>
        <CardDescription>{t('profile.avatarHint')}</CardDescription>
      </CardHeader>
      <CardContent className="flex flex-wrap items-center gap-4">
        <Avatar className="size-16">
          {avatarUrl && <AvatarImage src={avatarUrl} alt="" />}
          <AvatarFallback className="text-lg">{initials(label)}</AvatarFallback>
        </Avatar>
        <div className="flex flex-wrap items-center gap-2">
          <input
            ref={inputRef}
            type="file"
            accept="image/png,image/jpeg,image/webp"
            className="hidden"
            onChange={onFile}
          />
          <Button
            variant="outline"
            onClick={() => inputRef.current?.click()}
            disabled={avatar.isPending}
          >
            {t('profile.upload')}
          </Button>
          {avatarUrl && (
            <Button
              variant="ghost"
              onClick={() => avatar.mutate('')}
              disabled={avatar.isPending}
            >
              {t('profile.remove')}
            </Button>
          )}
        </div>
        {(error || avatar.error) && (
          <p className="w-full text-sm text-destructive">
            {error || (avatar.error as Error).message}
          </p>
        )}
      </CardContent>
    </Card>
  );
}

function AccountCard({ nickname, email }: { nickname: string; email: string }) {
  const { t } = useTranslation();
  const update = useUpdateProfile();
  const [nick, setNick] = useState(nickname);
  const [mail, setMail] = useState(email);

  const onSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    update.mutate({ nickname: nick, email: mail });
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t('profile.account')}</CardTitle>
      </CardHeader>
      <CardContent>
        <form onSubmit={onSubmit} className="max-w-sm space-y-4">
          <div className="space-y-1.5">
            <Label htmlFor="nickname">{t('profile.nickname')}</Label>
            <Input id="nickname" value={nick} onChange={(e) => setNick(e.target.value)} />
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="p-email">{t('profile.email')}</Label>
            <Input
              id="p-email"
              type="email"
              value={mail}
              onChange={(e) => setMail(e.target.value)}
            />
          </div>
          {update.error && (
            <p className="text-sm text-destructive">{(update.error as Error).message}</p>
          )}
          {update.isSuccess && !update.isPending && (
            <p className="text-sm text-green-600">{t('profile.saved')}</p>
          )}
          <Button type="submit" disabled={update.isPending}>
            {t('profile.saveChanges')}
          </Button>
        </form>
      </CardContent>
    </Card>
  );
}

function PasswordCard() {
  const { t } = useTranslation();
  const change = useChangePassword();
  const [current, setCurrent] = useState('');
  const [next, setNext] = useState('');

  const onSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    change.mutate(
      { currentPassword: current, newPassword: next },
      {
        onSuccess: () => {
          setCurrent('');
          setNext('');
        },
      },
    );
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t('profile.changePassword')}</CardTitle>
      </CardHeader>
      <CardContent>
        <form onSubmit={onSubmit} className="max-w-sm space-y-4">
          <div className="space-y-1.5">
            <Label htmlFor="cur-pw">{t('profile.currentPassword')}</Label>
            <Input
              id="cur-pw"
              type="password"
              autoComplete="current-password"
              value={current}
              onChange={(e) => setCurrent(e.target.value)}
            />
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="new-pw">{t('profile.newPassword')}</Label>
            <Input
              id="new-pw"
              type="password"
              autoComplete="new-password"
              minLength={8}
              value={next}
              onChange={(e) => setNext(e.target.value)}
            />
          </div>
          {change.error && (
            <p className="text-sm text-destructive">{(change.error as Error).message}</p>
          )}
          {change.isSuccess && (
            <p className="text-sm text-green-600">{t('profile.passwordUpdated')}</p>
          )}
          <Button type="submit" disabled={change.isPending}>
            {t('profile.updatePassword')}
          </Button>
        </form>
      </CardContent>
    </Card>
  );
}
