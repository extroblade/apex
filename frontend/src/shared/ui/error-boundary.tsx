import { Component, type ReactNode } from 'react';

import { devlog } from '@/shared/lib/dev';

interface Props {
  children: ReactNode;
  fallback?: ReactNode;
}

interface State {
  hasError: boolean;
}

/**
 * Catches render/runtime errors in its subtree and shows a fallback instead of a
 * blank white screen. `componentDidCatch` is the single place to wire an error
 * reporter (Sentry etc.) — currently it just devlogs. Wrap the page content so a
 * crashing page leaves the shell (header/nav/footer) intact.
 */
export class ErrorBoundary extends Component<Props, State> {
  state: State = { hasError: false };

  static getDerivedStateFromError(): State {
    return { hasError: true };
  }

  componentDidCatch(error: unknown, info: unknown) {
    // Last-resort log; a real error tracker hooks in here.
    devlog('[error-boundary]', error, info);
  }

  reset = () => this.setState({ hasError: false });

  render() {
    if (!this.state.hasError) return this.props.children;
    if (this.props.fallback) return this.props.fallback;

    return (
      <div role="alert" className="mx-auto max-w-md rounded-lg border p-6 text-center">
        <h2 className="text-lg font-semibold">Something went wrong</h2>
        <p className="mt-2 text-sm text-muted-foreground">
          An unexpected error occurred. Try reloading the page.
        </p>
        <button
          type="button"
          onClick={() => window.location.reload()}
          className="mt-4 inline-flex h-9 cursor-pointer items-center justify-center rounded-md bg-primary px-4 text-sm font-medium text-primary-foreground hover:bg-primary/90"
        >
          Reload
        </button>
      </div>
    );
  }
}
