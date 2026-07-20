-- Account lifecycle: email change (pending email) + per-token target email.
--
-- pending_email on users holds the candidate address while the user is
-- verifying it. The old email stays valid for login until the new one is
-- confirmed, so a typo'd change can't lock the user out. On confirm, the
-- pending email is promoted to email and pending_email is cleared.
ALTER TABLE users
    ADD COLUMN pending_email VARCHAR(255) NULL;

-- target_email on email_tokens carries the address a verify token was issued
-- for. For the welcome/resend flow it's NULL (verifying the current email);
-- for the email-change flow it's the new address. consumeToken returns it so
-- the confirm step knows whether to swap emails or just flip the verified flag.
ALTER TABLE email_tokens
    ADD COLUMN target_email VARCHAR(255) NULL;
