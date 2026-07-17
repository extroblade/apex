import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';

import { App } from './App';
import { queryClient } from '@/shared/api/query-client';
import { fetchAvailableLocales } from '@/shared/i18n';
import { fetchViewer, viewerKeys } from '@/entities/viewer';
import { fetchFeatures, featureKeys } from '@/entities/features';
import { fetchNav, navKeys } from '@/entities/nav';
import { setCookie, deleteCookie } from '@/shared/lib/cookies';
import '@/shared/i18n/config';
import './styles/globals.css';

// Cockpit dev-overlay: ?dev=KEY sets a developer cookie; ?dev=off clears it.
// Runs before React mounts so the first render already sees the cookie.
(function initDevCookie() {
  const p = new URLSearchParams(window.location.search);
  const dev = p.get('dev');
  if (dev !== null) {
    if (dev === 'off') {
      deleteCookie('developer');
    } else if (dev !== '') {
      setCookie('developer', dev);
    }
    // Strip the param from the URL so it doesn't persist on refresh.
    p.delete('dev');
    const qs = p.toString();
    const url = window.location.pathname + (qs ? '?' + qs : '') + window.location.hash;
    window.history.replaceState(null, '', url);
  }
})();

// Boot preloading: fire the requests every screen needs (session, feature
// flags, menu, locale manifest) before React renders, so the first paint already
// has data in the cache instead of cascading loaders. (This is an SPA — true SSR
// would require a server-rendering framework; see CLAUDE.md roadmap.)
void queryClient.prefetchQuery({ queryKey: viewerKeys.me, queryFn: fetchViewer });
void queryClient.prefetchQuery({ queryKey: featureKeys.all, queryFn: fetchFeatures });
void queryClient.prefetchQuery({ queryKey: navKeys.all, queryFn: fetchNav });
void fetchAvailableLocales();

const container = document.getElementById('root');
if (!container) {
  throw new Error('Root element #root not found');
}

createRoot(container).render(
  <StrictMode>
    <App />
  </StrictMode>,
);
