import { useMemo, useState } from 'react';
import { Controller, useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Wand2, Layers, X } from 'lucide-react';

import { useCars, useTracks } from '@/entities/planner';
import {
  useCreateSetup,
  useGenerateSetup,
  useGenerateSetupPack,
  type GeneratedVariant,
} from '@/entities/setups';
import { useTranslation } from '@/shared/i18n';
import { Button } from '@/shared/ui/button';
import { Input } from '@/shared/ui/input';
import { Label } from '@/shared/ui/label';
import { Textarea } from '@/shared/ui/textarea';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/shared/ui/select';

const schema = z.object({
  name: z.string().trim().min(1, 'setups.errName'),
  carId: z.number().int().positive('setups.errCar'),
  trackId: z.number().int().nonnegative(),
  notes: z.string().trim().max(500, 'setups.errNotes'),
  data: z.string().trim().min(1, 'setups.errData'),
  public: z.boolean(),
});

type FormValues = z.infer<typeof schema>;

export function SetupForm({ onCreated }: { onCreated?: () => void }) {
  const { t } = useTranslation();
  const cars = useCars();
  const tracks = useTracks();
  const create = useCreateSetup();

  // One entry per base track (dedup layouts); trackId 0 = baseline / no track.
  const trackOptions = useMemo(() => {
    const seen = new Map<string, number>();
    for (const tr of tracks.data ?? []) {
      if (!seen.has(tr.trackName)) seen.set(tr.trackName, tr.trackId);
    }
    return Array.from(seen, ([name, id]) => ({ id, name })).sort((a, b) =>
      a.name.localeCompare(b.name),
    );
  }, [tracks.data]);

  const {
    register,
    handleSubmit,
    control,
    reset,
    watch,
    setValue,
    formState: { errors },
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { name: '', carId: 0, trackId: 0, notes: '', data: '', public: false },
  });

  const generate = useGenerateSetup();
  const generatePack = useGenerateSetupPack();
  const carId = watch('carId');
  const trackId = watch('trackId');

  // The generated pack (skill × session variants) shown for review, if any.
  const [pack, setPack] = useState<GeneratedVariant[] | null>(null);
  const [savingAll, setSavingAll] = useState(false);

  const loadVariant = (v: { name: string; notes: string; data: string }) => {
    setValue('name', v.name, { shouldValidate: true });
    setValue('notes', v.notes, { shouldValidate: true });
    setValue('data', v.data, { shouldValidate: true });
  };

  const onGenerate = () =>
    generate.mutate({ carId, trackId }, { onSuccess: loadVariant });

  const onGeneratePack = () =>
    generatePack.mutate({ carId, trackId }, { onSuccess: setPack });

  // Save every variant in the pack (sequential; keeps whatever succeeded if one
  // fails). Public defaults off — the user shares individually afterwards.
  const onSaveAll = async () => {
    if (!pack) return;
    setSavingAll(true);
    try {
      for (const v of pack) {
        await create.mutateAsync({
          name: v.name,
          carId,
          trackId,
          notes: v.notes,
          data: v.data,
          public: false,
        });
      }
      setPack(null);
      reset();
      onCreated?.();
    } finally {
      setSavingAll(false);
    }
  };

  const onSubmit = handleSubmit((values) => {
    create.mutate(values, {
      onSuccess: () => {
        reset();
        onCreated?.();
      },
    });
  });

  return (
    <form onSubmit={onSubmit} className="space-y-4" noValidate>
      <div className="space-y-1.5">
        <Label htmlFor="setup-name">{t('setups.name')}</Label>
        <Input id="setup-name" aria-invalid={!!errors.name} {...register('name')} />
        {errors.name && (
          <p className="text-sm text-destructive">{t(errors.name.message ?? '')}</p>
        )}
      </div>

      <div className="grid gap-4 sm:grid-cols-2">
        <div className="space-y-1.5">
          <Label>{t('setups.car')}</Label>
          <Controller
            control={control}
            name="carId"
            render={({ field }) => (
              <Select
                value={field.value ? String(field.value) : ''}
                onValueChange={(v) => field.onChange(Number(v))}
              >
                <SelectTrigger aria-invalid={!!errors.carId}>
                  <SelectValue placeholder={t('setups.selectCar')} />
                </SelectTrigger>
                <SelectContent>
                  {(cars.data ?? []).map((c) => (
                    <SelectItem key={c.carId} value={String(c.carId)}>
                      {c.carName}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            )}
          />
          {errors.carId && (
            <p className="text-sm text-destructive">{t(errors.carId.message ?? '')}</p>
          )}
        </div>

        <div className="space-y-1.5">
          <Label>{t('setups.track')}</Label>
          <Controller
            control={control}
            name="trackId"
            render={({ field }) => (
              <Select
                value={String(field.value)}
                onValueChange={(v) => field.onChange(Number(v))}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="0">{t('setups.baseline')}</SelectItem>
                  {trackOptions.map((tr) => (
                    <SelectItem key={tr.id} value={String(tr.id)}>
                      {tr.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            )}
          />
        </div>
      </div>

      <div className="space-y-1.5">
        <Label htmlFor="setup-notes">{t('setups.notes')}</Label>
        <Textarea
          id="setup-notes"
          rows={2}
          aria-invalid={!!errors.notes}
          {...register('notes')}
        />
        {errors.notes && (
          <p className="text-sm text-destructive">{t(errors.notes.message ?? '')}</p>
        )}
      </div>

      <div className="space-y-1.5">
        <Label htmlFor="setup-data">{t('setups.data')}</Label>
        <Textarea
          id="setup-data"
          rows={6}
          className="font-mono text-xs"
          placeholder={t('setups.dataHint')}
          aria-invalid={!!errors.data}
          {...register('data')}
        />
        {errors.data && (
          <p className="text-sm text-destructive">{t(errors.data.message ?? '')}</p>
        )}
      </div>

      <label className="flex cursor-pointer items-center gap-2 text-sm">
        <input
          type="checkbox"
          className="size-4 cursor-pointer"
          {...register('public')}
        />
        {t('setups.makePublic')}
      </label>

      {create.error && (
        <p className="text-sm text-destructive">{(create.error as Error).message}</p>
      )}
      {generate.error && (
        <p className="text-sm text-destructive">{(generate.error as Error).message}</p>
      )}
      {generatePack.error && (
        <p className="text-sm text-destructive">
          {(generatePack.error as Error).message}
        </p>
      )}

      <div className="flex flex-wrap gap-2">
        <Button type="submit" disabled={create.isPending || savingAll}>
          {create.isPending ? t('common.loading') : t('setups.save')}
        </Button>
        <Button
          type="button"
          variant="outline"
          disabled={!carId || generate.isPending}
          title={carId ? t('setups.generateHint') : t('setups.selectCar')}
          onClick={onGenerate}
        >
          <Wand2 className="size-4" />
          {generate.isPending ? t('common.loading') : t('setups.generate')}
        </Button>
        <Button
          type="button"
          variant="outline"
          disabled={!carId || generatePack.isPending}
          title={carId ? t('setups.packHint') : t('setups.selectCar')}
          onClick={onGeneratePack}
        >
          <Layers className="size-4" />
          {generatePack.isPending ? t('common.loading') : t('setups.generatePack')}
        </Button>
      </div>

      {pack && pack.length > 0 && (
        <PackPanel
          pack={pack}
          savingAll={savingAll}
          onUse={loadVariant}
          onSaveAll={onSaveAll}
          onClose={() => setPack(null)}
        />
      )}
    </form>
  );
}

/** Review panel for a generated pack: one card per variant, with per-variant
 *  "Use" (load into the form) and a "Save all" convenience. */
function PackPanel({
  pack,
  savingAll,
  onUse,
  onSaveAll,
  onClose,
}: {
  pack: GeneratedVariant[];
  savingAll: boolean;
  onUse: (v: GeneratedVariant) => void;
  onSaveAll: () => void;
  onClose: () => void;
}) {
  const { t } = useTranslation();
  return (
    <div className="space-y-3 rounded-md border p-3">
      <div className="flex items-center justify-between gap-2">
        <h3 className="text-sm font-medium">{t('setups.packTitle')}</h3>
        <div className="flex items-center gap-2">
          <Button type="button" size="sm" disabled={savingAll} onClick={onSaveAll}>
            {savingAll ? t('common.loading') : t('setups.saveAll', { n: pack.length })}
          </Button>
          <Button
            type="button"
            size="icon"
            variant="ghost"
            aria-label={t('setups.packClose')}
            onClick={onClose}
          >
            <X className="size-4" />
          </Button>
        </div>
      </div>
      <ul className="grid gap-2 sm:grid-cols-2">
        {pack.map((v) => (
          <li
            key={`${v.skill}-${v.session}`}
            className="flex items-center justify-between gap-3 rounded-md border px-3 py-2"
          >
            <span className="min-w-0">
              <span className="block truncate text-sm font-medium">{v.label}</span>
              <span className="block truncate text-xs text-muted-foreground">
                {v.notes}
              </span>
            </span>
            <Button
              type="button"
              size="sm"
              variant="outline"
              disabled={savingAll}
              onClick={() => onUse(v)}
            >
              {t('setups.use')}
            </Button>
          </li>
        ))}
      </ul>
    </div>
  );
}
