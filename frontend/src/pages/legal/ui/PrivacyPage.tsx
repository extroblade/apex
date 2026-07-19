import { LegalPage } from './LegalPage';

export function PrivacyPage() {
  return (
    <LegalPage title="Privacy Policy" updated="July 2026">
      <p>
        This policy explains what data Apex collects, why, and your choices. We aim to
        collect only what the Service needs to work.
      </p>

      <h2>What we collect</h2>
      <p>
        <strong>Account:</strong> your email address and a securely hashed password.{' '}
        <strong>Your content:</strong> setups, goals, race plans, and garage selections
        you create. <strong>Technical:</strong> basic request logs (e.g. IP address,
        timestamps) needed for security and reliability.
      </p>

      <h2>Why we use it</h2>
      <p>
        To provide the Service (authenticate you, store your work), to keep it secure
        (rate limiting, abuse prevention), and — for subscribers — to manage billing.
      </p>

      <h2>Cookies</h2>
      <p>
        We use a single essential cookie to keep you signed in. If product analytics are
        enabled, an analytics provider (e.g. Yandex Metrica) may set its own cookies;
        analytics are off unless configured.
      </p>

      <h2>Third parties</h2>
      <p>
        Catalog reference data may be sourced from third parties; we do not sell your
        personal data. When you link an external account (e.g. iRacing via its official
        API), we access only what you authorize.
      </p>

      <h2>Retention</h2>
      <p>
        We keep your data while your account is active. You can request deletion, and we
        remove your account and associated content, except where we must retain records to
        comply with law.
      </p>

      <h2>Your rights</h2>
      <p>
        Subject to your jurisdiction, you may request access to, correction of, export of,
        or deletion of your personal data. Contact us to exercise these rights.
      </p>

      <h2>Security</h2>
      <p>
        Passwords are hashed, sessions are token-based and revoked on password change, and
        any linked third-party credentials are encrypted at rest. No system is perfectly
        secure, but we take reasonable measures to protect your data.
      </p>

      <h2>Changes &amp; contact</h2>
      <p>
        We may update this policy and will note the date above. Questions or requests?
        Contact us at the support address listed on our site.
      </p>
    </LegalPage>
  );
}
