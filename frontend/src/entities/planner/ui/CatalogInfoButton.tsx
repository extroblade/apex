import { Info } from 'lucide-react';

import { CatalogThumbnail } from '@/shared/ui/catalog-thumbnail';
import { Badge } from '@/shared/ui/badge';
import { categoryMeta } from '@/shared/lib/racing-categories';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/shared/ui/dialog';

/** One selectable layout of a track (e.g. "Grand Prix", "Nordschleife"). */
export interface TrackLayout {
  name: string;
  free?: boolean;
}

export interface CatalogInfo {
  name: string;
  category: string;
  description: string;
  sub?: string; // track config / "layout"
  layouts?: TrackLayout[]; // all configs of a track, shown grouped in the dialog
}

/** An info button that opens a dialog with the item's generated art + details. */
export function CatalogInfoButton({
  name,
  category,
  description,
  sub,
  layouts,
}: CatalogInfo) {
  const { label, Icon } = categoryMeta(category);
  return (
    <Dialog>
      <DialogTrigger
        className="inline-flex size-8 shrink-0 cursor-pointer items-center justify-center rounded-md text-muted-foreground outline-none hover:bg-accent hover:text-accent-foreground focus-visible:ring-2 focus-visible:ring-ring"
        aria-label={`Details for ${name}`}
      >
        <Info className="size-4" />
      </DialogTrigger>
      <DialogContent>
        <div className="flex items-start gap-4">
          <CatalogThumbnail category={category} name={name} className="size-16" />
          <DialogHeader>
            <DialogTitle>{name}</DialogTitle>
            <div className="flex flex-wrap items-center gap-2 pt-1">
              <Badge>
                <Icon className="size-3" />
                {label}
              </Badge>
              {sub && <span className="text-sm text-muted-foreground">{sub}</span>}
            </div>
          </DialogHeader>
        </div>
        <DialogDescription className="pt-4 text-sm leading-relaxed text-foreground">
          {description || 'No description available yet.'}
        </DialogDescription>
        {layouts && layouts.length > 0 && (
          <div className="pt-4">
            <p className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
              Layouts
            </p>
            <ul className="mt-2 flex flex-wrap gap-1.5">
              {layouts.map((l) => (
                <li key={l.name}>
                  <span
                    className={
                      l.free
                        ? 'inline-flex items-center gap-1 rounded-full bg-green-500/15 px-2 py-0.5 text-xs text-green-700 ring-1 ring-green-600/40'
                        : 'inline-flex items-center rounded-full bg-muted px-2 py-0.5 text-xs text-muted-foreground'
                    }
                  >
                    {l.name}
                    {l.free && ' · free'}
                  </span>
                </li>
              ))}
            </ul>
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}
