import { AppProviders } from './providers';
import { AppRouter } from './providers/router';
import { Header } from '@/widgets/header';
import { SideNav } from '@/widgets/side-nav';
import { BottomNav } from '@/widgets/bottom-nav';
import { Footer } from '@/widgets/footer';
import { Cockpit } from '@/features/cockpit';
import { ErrorBoundary } from '@/shared/ui/error-boundary';

export function App() {
  return (
    <AppProviders>
      {/* min-h-screen + flex-col so the footer sits at the bottom of the
          viewport on short pages instead of floating up under the content. */}
      <div className="flex min-h-screen flex-col bg-background">
        {/* a11y: lets keyboard users jump straight past the navigation. */}
        <a
          href="#main"
          className="sr-only focus:not-sr-only focus:absolute focus:left-4 focus:top-4 focus:z-50 focus:rounded-md focus:bg-primary focus:px-3 focus:py-2 focus:text-sm focus:text-primary-foreground"
        >
          Skip to content
        </a>
        <Header />
        {/* flex-1 so this row stretches to fill the viewport height; the
            sidebar and main both stretch with it. */}
        <div className="flex flex-1">
          {/* Desktop navigation; the bottom bar takes over below md. */}
          <SideNav />
          <main
            id="main"
            className="flex min-w-0 flex-1 flex-col px-4 py-8 pb-24 md:pb-8"
          >
            {/* flex-1 here grows to fill, pushing the footer down to the bottom. */}
            <div className="mx-auto flex w-full max-w-6xl flex-1 flex-col">
              <div className="flex-1">
                <ErrorBoundary>
                  <AppRouter />
                </ErrorBoundary>
              </div>
              <Footer />
            </div>
          </main>
        </div>
        <BottomNav />
        {/* Dev-only overlay; renders nothing without the `developer` cookie. */}
        <Cockpit />
      </div>
    </AppProviders>
  );
}
