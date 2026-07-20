import { Route, Switch } from 'wouter';

import { HomePage } from '@/pages/home';
import { FuelPage } from '@/pages/fuel';
import { LoginPage } from '@/pages/login';
import { ResetPasswordPage, ResetPasswordConfirmPage } from '@/pages/reset-password';
import { VerifyEmailPage } from '@/pages/verify-email';
import { DashboardPage } from '@/pages/dashboard';
import { ComparePage } from '@/pages/compare';
import { PlannerPage } from '@/pages/planner';
import { ThisWeekPage } from '@/pages/this-week';
import { GaragePage } from '@/pages/garage';
import { SetupsPage } from '@/pages/setups';
import { GoalsPage } from '@/pages/goals';
import { DriversPage } from '@/pages/drivers';
import { DriverProfilePage } from '@/pages/driver-profile';
import { ProfilePage } from '@/pages/profile';
import { AboutPage } from '@/pages/about';
import { TermsPage, PrivacyPage } from '@/pages/legal';

export function AppRouter() {
  return (
    <Switch>
      <Route path="/" component={HomePage} />
      <Route path="/fuel" component={FuelPage} />
      <Route path="/drivers" component={DriversPage} />
      <Route path="/drivers/:custId" component={DriverProfilePage} />
      <Route path="/planner" component={PlannerPage} />
      <Route path="/this-week" component={ThisWeekPage} />
      <Route path="/garage" component={GaragePage} />
      <Route path="/setups" component={SetupsPage} />
      <Route path="/goals" component={GoalsPage} />
      <Route path="/dashboard" component={DashboardPage} />
      <Route path="/compare" component={ComparePage} />
      <Route path="/login" component={LoginPage} />
      <Route path="/reset-password" component={ResetPasswordPage} />
      <Route path="/reset-password/confirm" component={ResetPasswordConfirmPage} />
      <Route path="/verify-email" component={VerifyEmailPage} />
      <Route path="/profile" component={ProfilePage} />
      <Route path="/about" component={AboutPage} />
      <Route path="/terms" component={TermsPage} />
      <Route path="/privacy" component={PrivacyPage} />
      <Route>
        <p className="text-sm text-muted-foreground">404 — Not found</p>
      </Route>
    </Switch>
  );
}
