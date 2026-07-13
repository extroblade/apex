import { Route, Switch } from 'wouter';

import { HomePage } from '@/pages/home';
import { FuelPage } from '@/pages/fuel';
import { LoginPage } from '@/pages/login';
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
      <Route path="/profile" component={ProfilePage} />
      <Route path="/about" component={AboutPage} />
      <Route>
        <p className="text-sm text-muted-foreground">404 — Not found</p>
      </Route>
    </Switch>
  );
}
