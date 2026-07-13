import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';

import { App } from './App';
import { queryClient } from '@/shared/api/query-client';
import { fetchAvailableLocales } from '@/shared/i18n';
import { fetchViewer, viewerKeys } from '@/entities/viewer';
import { fetchFeatures, featureKeys } from '@/entities/features';
import '@/shared/i18n/config';
import './styles/globals.css';

// Boot preloading: fire the requests every screen needs (session, feature
// flags, locale manifest) before React renders, so the first paint already has
// data in the cache instead of cascading loaders. (This is an SPA — true SSR
// would require a server-rendering framework; see CLAUDE.md roadmap.)
void queryClient.prefetchQuery({ queryKey: viewerKeys.me, queryFn: fetchViewer });
void queryClient.prefetchQuery({ queryKey: featureKeys.all, queryFn: fetchFeatures });
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
