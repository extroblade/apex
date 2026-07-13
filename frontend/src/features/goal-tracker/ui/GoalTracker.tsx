import { Check, Minus, Plus, Trash2 } from 'lucide-react';

import { useGoals, useUpdateGoal, useDeleteGoal, type Goal } from '@/entities/goals';
import { useTranslation } from '@/shared/i18n';
import { Button } from '@/shared/ui/button';
import { SkeletonRows } from '@/shared/ui/skeleton';
import { cn } from '@/shared/lib/utils';

export function GoalTracker() {
  const { t } = useTranslation();
  const goals = useGoals();

  if (goals.isLoading) return <SkeletonRows rows={4} />;
  if (goals.error) {
    return <p className="text-sm text-destructive">{(goals.error as Error).message}</p>;
  }
  if ((goals.data ?? []).length === 0) {
    return <p className="text-sm text-muted-foreground">{t('goals.empty')}</p>;
  }

  return (
    <ul className="space-y-3">
      {(goals.data ?? []).map((g) => (
        <GoalRow key={g.id} goal={g} />
      ))}
    </ul>
  );
}

function GoalRow({ goal }: { goal: Goal }) {
  const { t } = useTranslation();
  const update = useUpdateGoal();
  const del = useDeleteGoal();

  // Reuse the full-update endpoint for small changes: send the goal back with
  // one field changed. `done` is left undefined so the server re-derives it.
  const patch = (changes: Partial<Pick<Goal, 'current' | 'done'>>) =>
    update.mutate({
      id: goal.id,
      input: {
        title: goal.title,
        notes: goal.notes,
        unit: goal.unit,
        target: goal.target,
        current: changes.current ?? goal.current,
        done: changes.done,
        dueDate: goal.dueDate,
      },
    });

  const pct = Math.round(goal.progress * 100);

  return (
    <li className={cn('rounded-md border p-3', goal.done && 'bg-muted/40')}>
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <div className="flex items-center gap-2">
            <span
              className={cn(
                'font-medium',
                goal.done && 'text-muted-foreground line-through',
              )}
            >
              {goal.title}
            </span>
            {goal.done && (
              <span className="inline-flex items-center gap-1 rounded-full bg-green-500/15 px-2 py-0.5 text-xs text-green-700 ring-1 ring-green-600/40">
                <Check className="size-3" />
                {t('goals.done')}
              </span>
            )}
          </div>
          {goal.notes && (
            <p className="mt-0.5 text-sm text-muted-foreground">{goal.notes}</p>
          )}
          {goal.dueDate && (
            <p className="mt-0.5 text-xs text-muted-foreground">
              {t('goals.by', { date: goal.dueDate })}
            </p>
          )}
        </div>
        <div className="flex shrink-0 items-center gap-1">
          <Button
            size="sm"
            variant="outline"
            aria-pressed={goal.done}
            onClick={() => patch({ done: !goal.done })}
            aria-label={goal.done ? t('goals.reopen') : t('goals.markDone')}
            title={goal.done ? t('goals.reopen') : t('goals.markDone')}
          >
            <Check className="size-4" />
          </Button>
          <Button
            size="sm"
            variant="ghost"
            onClick={() => del.mutate(goal.id)}
            aria-label={t('goals.delete')}
            title={t('goals.delete')}
          >
            <Trash2 className="size-4 text-destructive" />
          </Button>
        </div>
      </div>

      {goal.target > 0 && (
        <div className="mt-3 space-y-1.5">
          <div className="flex items-center justify-between text-xs text-muted-foreground">
            <span>
              {formatNum(goal.current)} / {formatNum(goal.target)}
              {goal.unit ? ` ${goal.unit}` : ''}
            </span>
            <span>{pct}%</span>
          </div>
          <div className="h-2 w-full overflow-hidden rounded-full bg-muted">
            <div
              className={cn(
                'h-full rounded-full',
                goal.done ? 'bg-green-500' : 'bg-primary',
              )}
              style={{ width: `${pct}%` }}
            />
          </div>
          <div className="flex items-center gap-1 pt-1">
            <Button
              size="sm"
              variant="outline"
              onClick={() => patch({ current: Math.max(0, goal.current - 1) })}
              aria-label={t('goals.decrement')}
            >
              <Minus className="size-3.5" />
            </Button>
            <Button
              size="sm"
              variant="outline"
              onClick={() => patch({ current: goal.current + 1 })}
              aria-label={t('goals.increment')}
            >
              <Plus className="size-3.5" />
            </Button>
          </div>
        </div>
      )}
    </li>
  );
}

function formatNum(n: number): string {
  return Number.isInteger(n) ? String(n) : n.toFixed(1);
}
