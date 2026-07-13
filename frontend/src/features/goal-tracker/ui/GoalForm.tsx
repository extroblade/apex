import { Controller, useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';

import { useCreateGoal } from '@/entities/goals';
import { useTranslation } from '@/shared/i18n';
import { Button } from '@/shared/ui/button';
import { DatePicker } from '@/shared/ui/date-picker';
import { Input } from '@/shared/ui/input';
import { Label } from '@/shared/ui/label';
import { Textarea } from '@/shared/ui/textarea';

const schema = z.object({
  title: z.string().trim().min(1, 'goals.errTitle'),
  unit: z.string().trim().max(40),
  target: z.number('goals.errTarget').min(0, 'goals.errTarget'),
  current: z.number('goals.errCurrent').min(0, 'goals.errCurrent'),
  notes: z.string().trim().max(500),
  dueDate: z.string(),
});

type FormValues = z.infer<typeof schema>;

export function GoalForm({ onCreated }: { onCreated?: () => void }) {
  const { t } = useTranslation();
  const create = useCreateGoal();

  const {
    register,
    handleSubmit,
    control,
    reset,
    formState: { errors },
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { title: '', unit: '', target: 0, current: 0, notes: '', dueDate: '' },
  });

  const onSubmit = handleSubmit((v) => {
    create.mutate(
      { ...v, dueDate: v.dueDate || null },
      {
        onSuccess: () => {
          reset();
          onCreated?.();
        },
      },
    );
  });

  return (
    <form onSubmit={onSubmit} className="space-y-4" noValidate>
      <div className="space-y-1.5">
        <Label htmlFor="goal-title">{t('goals.goalTitle')}</Label>
        <Input
          id="goal-title"
          placeholder={t('goals.titlePlaceholder')}
          aria-invalid={!!errors.title}
          {...register('title')}
        />
        {errors.title && (
          <p className="text-sm text-destructive">{t(errors.title.message ?? '')}</p>
        )}
      </div>

      <div className="grid gap-4 sm:grid-cols-3">
        <div className="space-y-1.5">
          <Label htmlFor="goal-target">{t('goals.target')}</Label>
          <Input
            id="goal-target"
            type="number"
            min="0"
            step="any"
            aria-invalid={!!errors.target}
            {...register('target', { valueAsNumber: true })}
          />
        </div>
        <div className="space-y-1.5">
          <Label htmlFor="goal-current">{t('goals.current')}</Label>
          <Input
            id="goal-current"
            type="number"
            min="0"
            step="any"
            aria-invalid={!!errors.current}
            {...register('current', { valueAsNumber: true })}
          />
        </div>
        <div className="space-y-1.5">
          <Label htmlFor="goal-unit">{t('goals.unit')}</Label>
          <Input
            id="goal-unit"
            placeholder={t('goals.unitPlaceholder')}
            {...register('unit')}
          />
        </div>
      </div>

      <div className="space-y-1.5">
        <Label htmlFor="goal-due">{t('goals.dueDate')}</Label>
        <Controller
          control={control}
          name="dueDate"
          render={({ field }) => (
            <DatePicker id="goal-due" value={field.value} onChange={field.onChange} />
          )}
        />
      </div>

      <div className="space-y-1.5">
        <Label htmlFor="goal-notes">{t('goals.notes')}</Label>
        <Textarea id="goal-notes" rows={2} {...register('notes')} />
      </div>

      {create.error && (
        <p className="text-sm text-destructive">{(create.error as Error).message}</p>
      )}

      <Button type="submit" disabled={create.isPending}>
        {create.isPending ? t('common.loading') : t('goals.save')}
      </Button>
    </form>
  );
}
