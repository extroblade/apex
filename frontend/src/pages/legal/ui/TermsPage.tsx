import { LegalPage } from './LegalPage';

export function TermsPage() {
  return (
    <LegalPage title="Terms of Service" updated="July 2026">
      <p>
        These Terms govern your use of Apex (the “Service”). By creating an account or
        using the Service you agree to them. If you do not agree, do not use the Service.
      </p>

      <h2>The Service</h2>
      <p>
        Apex is a companion app for sim racers: a fuel &amp; stint calculator, setup
        generator, season planner, garage, and goal tracking. Some features are offered on
        a free tier and others as part of a paid subscription.
      </p>

      <h2>Accounts</h2>
      <p>
        You are responsible for the activity under your account and for keeping your
        password confidential. Provide accurate information and notify us of any
        unauthorized use. You must be old enough to form a binding contract in your
        jurisdiction.
      </p>

      <h2>Your content</h2>
      <p>
        You retain ownership of the setups, goals, plans, and other content you create. By
        marking content as public (e.g. publishing a setup to the showroom) you grant us a
        non-exclusive license to host and display it to other users. You are responsible
        for the content you upload.
      </p>

      <h2>Acceptable use</h2>
      <p>
        Do not misuse the Service: no unlawful activity, no attempts to break, overload,
        or reverse-engineer it, and no infringement of others’ rights.
      </p>

      <h2>Third-party platforms</h2>
      <p>
        Apex is an independent project and is not affiliated with, endorsed by, or
        sponsored by iRacing.com Motorsport Simulations, LLC. Any third-party names are
        used only to describe compatibility.
      </p>

      <h2>Subscriptions</h2>
      <p>
        Paid plans renew until cancelled. Prices and included features may change; we will
        give notice of material changes. Except where required by law, fees already paid
        are non-refundable.
      </p>

      <h2>No warranty &amp; liability</h2>
      <p>
        The Service is provided “as is”, without warranties of any kind. To the maximum
        extent permitted by law, we are not liable for indirect or consequential damages,
        and our total liability is limited to the amount you paid in the twelve months
        before the claim.
      </p>

      <h2>Changes &amp; contact</h2>
      <p>
        We may update these Terms; continued use after changes means you accept them.
        Questions? Contact us at the support address listed on our site.
      </p>
    </LegalPage>
  );
}
