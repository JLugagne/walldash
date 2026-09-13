# 6. First-device auto-owner, invitations and rescue mode

Date: 2026-09-13

## Status

Accepted

Supersedes the OTP design documented in `docs/auth-custom-otp.md`.

## Context

The original device authentication flow issued a per-device one-time code (OTP) and wrote the
plaintext code to the server log (`otp_issued`). That log line was the only bootstrap channel for
the very first device, before any owner existed to approve it. In practice this made onboarding
awkward: the operator had to read a code out of the add-on log for the first tablet, and every
later device still depended on a short-lived code.

A secret in the logs is also a liability: anyone with access to the add-on log or a log export
could enroll a device.

## Decision

Replace OTP onboarding with three mechanisms:

1. **First-run bootstrap.** When the `auth_accounts` table is empty, the first device to call
   `POST /api/auth/connect` is created as `owner` and signed in immediately. No secret is logged.
2. **Owner approval.** Once any account exists, a device that connects with no invitation becomes
   a pending enrollment. An owner or admin approves or denies it from **Setup → Access**; the
   waiting device redeems its opaque pending cookie and signs in as `device`.
3. **Invitations.** An owner or admin mints a single-use invitation (role `device` or `admin`,
   TTL 15 minutes, revocable). The plaintext token/link is returned exactly once and only its
   selector + verifier hash are stored, reusing egauth's exported verification-token primitives.
   Opening `…/?invite=<token>` signs the device in.

A `rescue_mode` add-on option (env `RESCUE_MODE`) handles a lost owner device: while enabled, the
next unauthenticated device to connect claims the `owner` role even though accounts exist. The
mode is consumed after that single use (in memory) and re-armed by a restart, so it is effectively
one-shot per process. The server logs a warning at startup and whenever the mode is used.

The OTP service and the `/api/auth/verify` endpoint are removed. The access/refresh token model is
unchanged: the access token carries the account's roles and scopes (`setup:manage` for
owner/admin), and refresh rotation re-reads the account.

## Consequences

- A fresh instance is owned by whoever opens it first. It must be deployed on a trusted network
  and set up promptly; this is the deliberate trade-off for removing the log-delivered secret.
- Device onboarding no longer requires reading logs. The first device is instant; later devices
  are one click (approve) or one link (invite).
- Invitations are persisted in `auth_invites`, so they survive restarts, are revocable and are
  auditable. Pending enrollments and rescue consumption stay in memory (single-replica only, as
  before).
- `rescue_mode` left enabled across a restart re-arms; documentation and startup/consumption
  warnings tell the operator to set it back to `false`.
- The owner role can no longer be granted through an invitation; only bootstrap and rescue create
  an owner. Admins can never escalate to owner.
