import { AppProviders } from './providers';
import { AppRouter } from './providers/router';
import { Header } from '@/widgets/header';
import { BottomNav } from '@/widgets/bottom-nav';
import { Cockpit } from '@/features/cockpit';

export function App() {
  return (
    <AppProviders>
      <div className="min-h-screen bg-background">
        {/* a11y: lets keyboard users jump straight past the navigation. */}
        <a
          href="#main"
          className="sr-only focus:not-sr-only focus:absolute focus:left-4 focus:top-4 focus:z-50 focus:rounded-md focus:bg-primary focus:px-3 focus:py-2 focus:text-sm focus:text-primary-foreground"
        >
          Skip to content
        </a>
        <Header />
        <main id="main" className="mx-auto max-w-6xl px-4 py-8 pb-24 md:pb-8">
          <AppRouter />
        </main>
        <BottomNav />
        {/* Dev-only overlay; renders nothing without the `developer` cookie. */}
        <Cockpit />
      </div>
    </AppProviders>
  );
}
