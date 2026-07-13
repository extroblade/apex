import { useEffect } from 'react';
import { useLocation } from 'wouter';

import { AuthForm } from '@/features/auth';
import { useViewer } from '@/entities/viewer';
import { Aurora } from '@/shared/ui/fx/aurora';

export function LoginPage() {
  const { data: viewer } = useViewer();
  const [, navigate] = useLocation();

  // Already logged in? Bounce to the home page.
  useEffect(() => {
    if (viewer) navigate('/');
  }, [viewer, navigate]);

  return (
    <div className="relative py-8">
      <Aurora />
      <div className="relative">
        <AuthForm />
      </div>
    </div>
  );
}
