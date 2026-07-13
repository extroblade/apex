import { useState } from 'react';
import { Download, Globe, Lock, Trash2, User } from 'lucide-react';

import {
  useSetups,
  useSetup,
  useSetSetupPublic,
  useDeleteSetup,
  type Setup,
} from '@/entities/setups';
import { useTranslation } from '@/shared/i18n';
import { Button } from '@/shared/ui/button';
import { Badge } from '@/shared/ui/badge';
import { CatalogThumbnail } from '@/shared/ui/catalog-thumbnail';
import { SkeletonRows } from '@/shared/ui/skeleton';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/shared/ui/dialog';
import { Card, CardContent } from '@/shared/ui/card';
import { cn } from '@/shared/lib/utils';

export function SetupsShowroom() {
  const { t } = useTranslation();
  const [mine, setMine] = useState(false);
  const setups = useSetups(mine);
  const [openId, setOpenId] = useState<number | null>(null);

  return (
    <div className="space-y-4">
      <div className="flex gap-2" role="group" aria-label={t('setups.title')}>
        <Button
          size="sm"
          variant={mine ? 'outline' : 'default'}
          aria-pressed={!mine}
          onClick={() => setMine(false)}
        >
          {t('setups.showroom')}
        </Button>
        <Button
          size="sm"
          variant={mine ? 'default' : 'outline'}
          aria-pressed={mine}
          onClick={() => setMine(true)}
        >
          {t('setups.mine')}
        </Button>
      </div>

      {setups.isLoading ? (
        <SkeletonRows rows={4} />
      ) : setups.error ? (
        <p className="text-sm text-destructive">{(setups.error as Error).message}</p>
      ) : (setups.data ?? []).length === 0 ? (
        <p className="text-sm text-muted-foreground">{t('setups.empty')}</p>
      ) : (
        <ul className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {(setups.data ?? []).map((s) => (
            <SetupCard key={s.id} setup={s} onOpen={() => setOpenId(s.id)} />
          ))}
        </ul>
      )}

      <SetupDialog id={openId} onClose={() => setOpenId(null)} />
    </div>
  );
}

function SetupCard({ setup, onOpen }: { setup: Setup; onOpen: () => void }) {
  const { t } = useTranslation();
  const setPublic = useSetSetupPublic();
  const del = useDeleteSetup();

  return (
    <li className="flex flex-col gap-2 rounded-md border p-3">
      <div className="flex items-start gap-3">
        <CatalogThumbnail category={setup.category} className="size-10" />
        <div className="min-w-0 flex-1">
          <button
            type="button"
            onClick={onOpen}
            className="block w-full cursor-pointer truncate text-left text-sm font-medium hover:underline"
          >
            {setup.name}
          </button>
          <div className="truncate text-xs text-muted-foreground">
            {setup.carName}
            {setup.trackId > 0 ? ` · ${setup.trackName}` : ` · ${t('setups.baseline')}`}
          </div>
        </div>
        <Badge className={cn(setup.public ? 'text-green-600' : 'text-muted-foreground')}>
          {setup.public ? <Globe className="size-3" /> : <Lock className="size-3" />}
        </Badge>
      </div>

      {setup.notes && (
        <p className="line-clamp-2 text-xs text-muted-foreground">{setup.notes}</p>
      )}

      <div className="mt-auto flex items-center justify-between pt-1 text-xs text-muted-foreground">
        <span className="inline-flex items-center gap-1">
          <User className="size-3" />
          {setup.mine ? t('setups.you') : setup.author}
        </span>
        <span className="inline-flex items-center gap-2">
          <span className="inline-flex items-center gap-1">
            <Download className="size-3" />
            {setup.downloads}
          </span>
          {setup.mine && (
            <>
              <button
                type="button"
                onClick={() => setPublic.mutate({ id: setup.id, public: !setup.public })}
                className="cursor-pointer hover:text-foreground"
                aria-label={setup.public ? t('setups.unpublish') : t('setups.publish')}
                title={setup.public ? t('setups.unpublish') : t('setups.publish')}
              >
                {setup.public ? (
                  <Lock className="size-3.5" />
                ) : (
                  <Globe className="size-3.5" />
                )}
              </button>
              <button
                type="button"
                onClick={() => del.mutate(setup.id)}
                className="cursor-pointer text-destructive hover:opacity-80"
                aria-label={t('setups.delete')}
                title={t('setups.delete')}
              >
                <Trash2 className="size-3.5" />
              </button>
            </>
          )}
        </span>
      </div>
    </li>
  );
}

/** Opening the dialog fetches the setup with ?download=1 to bump the counter. */
function SetupDialog({ id, onClose }: { id: number | null; onClose: () => void }) {
  const { t } = useTranslation();
  const setup = useSetup(id, true);
  const s = setup.data;

  return (
    <Dialog open={id != null} onOpenChange={(open) => !open && onClose()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{s ? s.name : t('common.loading')}</DialogTitle>
        </DialogHeader>

        {setup.isLoading && <SkeletonRows rows={4} />}

        {s && (
          <div className="space-y-3">
            <div className="text-sm text-muted-foreground">
              {s.carName}
              {s.trackId > 0 ? ` · ${s.trackName}` : ` · ${t('setups.baseline')}`}
              {' · '}
              {s.mine ? t('setups.you') : s.author}
            </div>
            {s.notes && <p className="text-sm">{s.notes}</p>}
            <Card>
              <CardContent className="pt-4">
                <pre className="max-h-72 overflow-auto whitespace-pre-wrap break-words font-mono text-xs">
                  {s.data}
                </pre>
              </CardContent>
            </Card>
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}
