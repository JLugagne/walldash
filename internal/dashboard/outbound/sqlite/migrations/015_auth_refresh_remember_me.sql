-- egauth v0.16 records two more facts on the refresh family so a silent rotation carries them
-- onto the descendant and its claims:
--
--   * remember_me: the cookie persistence choice made at login. Rotate replays it verbatim, so a
--     session cookie stays a session cookie and a "remember me" cookie stays persistent across
--     every rotation (before v0.16 the rotation downgraded a persistent cookie to a session one).
--   * amr: the authentication method references proved when the family was minted. Rotate falls
--     back to it when the ClaimsProvider returns no AMR, so a stepped-up session keeps its
--     assurance level across a silent refresh instead of silently decaying.
--
-- Legacy rows read back as 0 / '' (session cookie, no proved assurance), which matches their
-- previous meaning. amr is stored as a JSON array of strings.
ALTER TABLE auth_refresh_tokens ADD COLUMN remember_me INTEGER NOT NULL DEFAULT 0;
ALTER TABLE auth_refresh_tokens ADD COLUMN amr TEXT NOT NULL DEFAULT '';
