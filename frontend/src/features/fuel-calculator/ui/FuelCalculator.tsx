import { useState } from 'react';
import { Info } from 'lucide-react';

import { useTranslation } from '@/shared/i18n';
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
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/shared/ui/dialog';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/shared/ui/select';
import { useFuelPlan } from '../api/use-fuel-plan';
import type { RaceType, FuelStrategy, PitWindow, WindowUnit } from '../model/types';

const STRATEGIES: FuelStrategy[] = ['balanced', 'undercut', 'overcut'];

const STRATEGY_LABEL: Record<FuelStrategy, string> = {
  balanced: 'fuel.strategyBalanced',
  undercut: 'fuel.strategyUndercut',
  overcut: 'fuel.strategyOvercut',
};

// Keep the rules UI manageable: windows are configurable for up to 4 stops.
const MAX_WINDOW_ROWS = 4;

export function FuelCalculator() {
  const { t } = useTranslation();
  const [raceType, setRaceType] = useState<RaceType>('laps');
  const [strategy, setStrategy] = useState<FuelStrategy>('balanced');
  const [raceLaps, setRaceLaps] = useState('30');
  const [raceMinutes, setRaceMinutes] = useState('30');
  const [lapTimeSec, setLapTimeSec] = useState('90');
  const [fuelPerLap, setFuelPerLap] = useState('2.5');
  const [tankCapacity, setTankCapacity] = useState('50');
  const [extraLaps, setExtraLaps] = useState('1');
  const [mandatoryStops, setMandatoryStops] = useState('0');
  const [windows, setWindows] = useState<PitWindow[]>([]);

  const { mutate, data: plan, error, isPending } = useFuelPlan();

  const stops = Math.max(0, Number(mandatoryStops) || 0);
  const windowRows = Math.min(stops, MAX_WINDOW_ROWS);
  // Minute-based windows need the average lap time even for lap races.
  const needLapTime =
    raceType === 'time' ||
    windows
      .slice(0, windowRows)
      .some((w) => w.unit === 'minutes' && (w.from > 0 || w.to > 0));

  const setWindow = (i: number, patch: Partial<PitWindow>) => {
    setWindows((prev) => {
      const next = [...prev];
      while (next.length <= i) next.push({ from: 0, to: 0, unit: 'laps' });
      next[i] = { ...next[i], ...patch };
      return next;
    });
  };

  const onSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    mutate({
      raceType,
      strategy,
      raceLaps: Number(raceLaps),
      raceMinutes: Number(raceMinutes),
      lapTimeSec: Number(lapTimeSec),
      fuelPerLap: Number(fuelPerLap),
      tankCapacity: Number(tankCapacity),
      extraLaps: Number(extraLaps),
      rules: {
        mandatoryStops: stops,
        windows: windows
          .slice(0, windowRows)
          .map((w) => ({ ...w, from: Number(w.from) || 0, to: Number(w.to) || 0 })),
      },
    });
  };

  return (
    <div className="grid gap-6 md:grid-cols-2">
      <Card>
        <CardHeader>
          <CardTitle>{t('fuel.calcTitle')}</CardTitle>
          <CardDescription>{t('fuel.calcDesc')}</CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={onSubmit} className="space-y-4">
            <div className="flex gap-2" role="group" aria-label={t('fuel.calcTitle')}>
              <Button
                type="button"
                size="sm"
                variant={raceType === 'laps' ? 'default' : 'outline'}
                onClick={() => setRaceType('laps')}
              >
                {t('fuel.byLaps')}
              </Button>
              <Button
                type="button"
                size="sm"
                variant={raceType === 'time' ? 'default' : 'outline'}
                onClick={() => setRaceType('time')}
              >
                {t('fuel.byTime')}
              </Button>
            </div>

            {raceType === 'laps' ? (
              <Field label={t('fuel.raceLaps')} value={raceLaps} onChange={setRaceLaps} />
            ) : (
              <Field
                label={t('fuel.raceLength')}
                value={raceMinutes}
                onChange={setRaceMinutes}
              />
            )}
            {needLapTime && (
              <Field
                label={t('fuel.avgLapTime')}
                value={lapTimeSec}
                onChange={setLapTimeSec}
              />
            )}

            <Field
              label={t('fuel.fuelPerLap')}
              value={fuelPerLap}
              onChange={setFuelPerLap}
            />
            <Field
              label={t('fuel.tank')}
              value={tankCapacity}
              onChange={setTankCapacity}
            />
            <Field label={t('fuel.margin')} value={extraLaps} onChange={setExtraLaps} />

            <div className="space-y-1.5">
              <div className="flex items-center gap-1">
                <Label>{t('fuel.strategy')}</Label>
                <StrategyInfo />
              </div>
              <div
                className="flex flex-wrap gap-2"
                role="group"
                aria-label={t('fuel.strategy')}
              >
                {STRATEGIES.map((s) => (
                  <Button
                    key={s}
                    type="button"
                    size="sm"
                    variant={strategy === s ? 'default' : 'outline'}
                    aria-pressed={strategy === s}
                    onClick={() => setStrategy(s)}
                  >
                    {t(STRATEGY_LABEL[s])}
                  </Button>
                ))}
              </div>
            </div>

            <fieldset className="space-y-3 rounded-md border p-3">
              <legend className="px-1 text-sm font-medium">{t('fuel.rulesTitle')}</legend>
              <p className="text-xs text-muted-foreground">{t('fuel.rulesDesc')}</p>
              <div className="space-y-1.5">
                <Label htmlFor="mand-stops">{t('fuel.mandatoryStops')}</Label>
                <Input
                  id="mand-stops"
                  type="number"
                  min="0"
                  max="10"
                  value={mandatoryStops}
                  onChange={(e) => setMandatoryStops(e.target.value)}
                />
              </div>

              {Array.from({ length: windowRows }).map((_, i) => {
                const w = windows[i] ?? { from: 0, to: 0, unit: 'laps' as WindowUnit };
                return (
                  <div key={i} className="space-y-1.5">
                    <Label>{t('fuel.windowFor', { n: i + 1 })}</Label>
                    <div className="flex flex-wrap items-center gap-2">
                      <Input
                        type="number"
                        min="0"
                        placeholder={t('fuel.windowFrom')}
                        aria-label={`${t('fuel.windowFor', { n: i + 1 })} — ${t('fuel.windowFrom')}`}
                        value={w.from || ''}
                        onChange={(e) => setWindow(i, { from: Number(e.target.value) })}
                        className="w-24"
                      />
                      <span className="text-muted-foreground">–</span>
                      <Input
                        type="number"
                        min="0"
                        placeholder={t('fuel.windowTo')}
                        aria-label={`${t('fuel.windowFor', { n: i + 1 })} — ${t('fuel.windowTo')}`}
                        value={w.to || ''}
                        onChange={(e) => setWindow(i, { to: Number(e.target.value) })}
                        className="w-24"
                      />
                      <Select
                        value={w.unit}
                        onValueChange={(v) => setWindow(i, { unit: v as WindowUnit })}
                      >
                        <SelectTrigger
                          className="w-28"
                          aria-label={`${t('fuel.windowFor', { n: i + 1 })} — unit`}
                        >
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="laps">{t('fuel.unitLaps')}</SelectItem>
                          <SelectItem value="minutes">{t('fuel.unitMinutes')}</SelectItem>
                        </SelectContent>
                      </Select>
                    </div>
                  </div>
                );
              })}
            </fieldset>

            <Button type="submit" className="w-full" disabled={isPending}>
              {isPending ? t('fuel.calculating') : t('fuel.calculate')}
            </Button>
          </form>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>{t('fuel.planTitle')}</CardTitle>
          <CardDescription>{t('fuel.planDesc')}</CardDescription>
        </CardHeader>
        <CardContent>
          {error && (
            <p className="text-sm text-destructive">{(error as Error).message}</p>
          )}
          {!plan && !error && (
            <p className="text-sm text-muted-foreground">{t('fuel.fillForm')}</p>
          )}
          {plan && (
            <div className="space-y-4">
              <div className="grid grid-cols-2 gap-3">
                <Stat label={t('fuel.totalLaps')} value={plan.totalLaps} />
                <Stat label={t('fuel.pitStops')} value={plan.pitStops} />
                <Stat label={t('fuel.lapsPerStint')} value={plan.lapsPerStint} />
                <Stat label={t('fuel.startFuel')} value={`${plan.startFuel} L`} />
                <Stat label={t('fuel.fuelNeeded')} value={`${plan.fuelNeeded} L`} />
                <Stat label={t('fuel.withMargin')} value={`${plan.fuelWithMargin} L`} />
              </div>
              <div>
                <p className="mb-2 text-sm font-medium">{t('fuel.stints')}</p>
                <ul className="text-sm">
                  {plan.stints.map((s) => (
                    <li
                      key={s.index}
                      className="flex items-center justify-between gap-2 border-b py-1.5 last:border-0"
                    >
                      <span className="text-muted-foreground">
                        {t('fuel.stint')} {s.index}
                      </span>
                      <span>
                        {s.laps} {t('fuel.laps')}
                        {s.pitOnLap > 0 && (
                          <span className="text-muted-foreground">
                            {' '}
                            · {t('fuel.pitOnLap', { lap: s.pitOnLap })}
                          </span>
                        )}
                      </span>
                      <span className="font-medium">{s.fuel} L</span>
                    </li>
                  ))}
                </ul>
              </div>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

/** Info dialog explaining when each strategy is the right call. */
function StrategyInfo() {
  const { t } = useTranslation();
  return (
    <Dialog>
      <DialogTrigger
        className="inline-flex size-6 cursor-pointer items-center justify-center rounded-md text-muted-foreground outline-none hover:bg-accent hover:text-accent-foreground focus-visible:ring-2 focus-visible:ring-ring"
        aria-label={t('fuel.strategyInfoTitle')}
      >
        <Info className="size-4" />
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{t('fuel.strategyInfoTitle')}</DialogTitle>
        </DialogHeader>
        <div className="space-y-3 pt-2 text-sm leading-relaxed">
          <p>{t('fuel.strategyInfoBalanced')}</p>
          <p>{t('fuel.strategyInfoUndercut')}</p>
          <p>{t('fuel.strategyInfoOvercut')}</p>
        </div>
      </DialogContent>
    </Dialog>
  );
}

function Field({
  label,
  value,
  onChange,
}: {
  label: string;
  value: string;
  onChange: (v: string) => void;
}) {
  return (
    <div className="space-y-1.5">
      <Label>{label}</Label>
      <Input
        type="number"
        inputMode="decimal"
        step="any"
        min="0"
        value={value}
        onChange={(e) => onChange(e.target.value)}
      />
    </div>
  );
}

function Stat({ label, value }: { label: string; value: React.ReactNode }) {
  return (
    <div className="rounded-md border p-2">
      <div className="text-xs text-muted-foreground">{label}</div>
      <div className="text-base font-semibold">{value}</div>
    </div>
  );
}
