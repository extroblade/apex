# Apex -> ContentPilot Transition Plan

Status: approved  
Date: 2026-07-20  
Current repo: `apex`  
New repo target: `git@github.com:extroblade/ContentPilot.git`  
Chosen product name: **ContentPilot**

---

## 1) Goal

1. Freeze Apex in a stable, working state.
2. Start a new private repository as **ContentPilot**.
3. Move the technical platform/base code into the new repo.
4. Ensure all new commits are authored by:
   - `extroblade <extroblade@users.noreply.github.com>`
5. Continue development in a separate chat/context.

---

## 2) Apex closure checklist (this repo)

- [x] Main branch includes latest billing + CI fixes.
- [x] CI green on latest changes before freeze.
- [ ] No new product features for Apex (only emergency fixes if needed).
- [ ] Optional: tag final Apex baseline (e.g. `apex-final-v1`).
- [ ] Optional: add short note in README that active development moved to ContentPilot.

Decision: Apex remains as historical baseline; new product work moves to new repo.

---

## 3) Contributor cleanup strategy for new repo

Requirement: "prefer to clean other contributors."

### Recommended method (cleanest): **orphan snapshot import**

This keeps code/files but drops git history, producing a fresh history with your author only.

### Why this method

- No mixed authors in commit history.
- No need to rewrite old public history in Apex.
- Fast and low-risk for a private reboot.

---

## 4) Exact migration steps to ContentPilot

Run from local clone of current `apex`:

```bash
# 0) make sure apex main is up to date
git checkout main
git pull origin main

# 1) create a working copy for migration
cd ..
git clone /workspace ContentPilot
cd ContentPilot

# 2) point to new remote
git remote remove origin
git remote add origin git@github.com:extroblade/ContentPilot.git

# 3) create orphan branch (no previous history)
git checkout --orphan main

# 4) clear git index and re-add all files
git rm -rf --cached .
git add .

# 5) ensure author identity
git config user.name "extroblade"
git config user.email "extroblade@users.noreply.github.com"

# 6) first clean root commit
git commit -m "chore: initialize ContentPilot from Apex baseline"

# 7) push as fresh main
git push -u origin main --force
```

Result: new repository starts with one commit authored by you.

---

## 5) What to remove/rename immediately in ContentPilot

Before active feature development:

1. Rename brand references:
   - `Apex` -> `ContentPilot`
   - logos/favicons/text labels.
2. Remove iRacing-specific domain:
   - routes, entities, docs, labels, seed data tied to iRacing.
3. Keep platform modules:
   - auth/lifecycle
   - billing skeleton
   - feature flags
   - i18n infrastructure
   - CI and deployment skeleton
4. Replace navigation with content-ops sections:
   - Ideas
   - Episodes
   - Pipeline
   - Distribution
   - Metrics
   - Settings/Billing

---

## 6) Commit policy for ContentPilot

Always enforce in repo-local config:

```bash
git config user.name "extroblade"
git config user.email "extroblade@users.noreply.github.com"
```

For safety, verify before every push:

```bash
git config --get user.name
git config --get user.email
```

---

## 7) Phase-1 build scope in new chat

First implementation phase (suggested):

1. Rebrand + remove iRacing domain.
2. Create core entities:
   - Channel
   - Idea
   - Episode
   - Derivative
   - PublishTarget
3. Implement first usable workflow:
   - Idea -> Script -> Publish -> Repurpose.
4. Add weekly dashboard and status board.
5. Keep Stripe billing foundation but disable paywall until core workflow is stable.

---

## 8) Risks and controls

1. **Risk:** carrying too much Apex domain baggage.  
   **Control:** explicit delete list before feature coding.

2. **Risk:** rebuilding too broad too early.  
   **Control:** ship one complete pipeline first (long-video -> derivatives).

3. **Risk:** hidden author mix in new repo.  
   **Control:** orphan import + single root commit + pre-push author check.

---

## 9) Definition of transition complete

Transition is complete when:

- ContentPilot repo exists and is private.
- Fresh `main` history starts with your authored root commit.
- Baseline builds/tests pass in new repo.
- New chat starts from ContentPilot plan and implementation scope.

