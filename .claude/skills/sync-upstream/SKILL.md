---
name: sync-upstream
description: Update this tun2socks fork from upstream (xjasonlyu/tun2socks) up to a given tag/version. Diffs from the fork's actual divergence point (not a naive merge), skips changes that only undo this fork's rebranding, and adapts CI/README changes to this fork's own workflows. Triggers on "update from upstream", "sync/merge upstream vX.Y.Z", "pull in the new tun2socks release", "what's new upstream".
---

# Sync this fork from upstream tun2socks

This repo (`github.com/Stanislav-Povolotsky/tun2socks`) is a rebranded, feature-diverged fork
of `github.com/xjasonlyu/tun2socks`. This skill brings in upstream changes up to a requested
version without blindly reverting this fork's own identity or re-triggering conflicts with
features upstream doesn't have.

**Usage**: `/sync-upstream [version]` — e.g. `/sync-upstream v2.7.0`. If no version is given,
ask the user which upstream tag to sync to (don't assume "latest" silently — the user tracks
this deliberately, one version at a time).

## Why this needs a skill instead of a plain `git merge`

- **This fork has been mechanically renamed.** Module path, README badges/links, Dockerfile
  label, `.golangci.yaml` import prefix, Docker Hub/GHCR tags, issue template links, and
  `SECURITY.md`/`CODE_OF_CONDUCT.md` contact addresses were all changed from `xjasonlyu` to
  `Stanislav-Povolotsky`. A plain `git merge upstream/main` would try to revert every one of
  these on every sync. Diff against the actual fork point instead (see below), and treat any
  upstream hunk that's *purely* a rename/branding change as "ours wins" — but still read it,
  because sometimes a branding-adjacent commit also carries real content (e.g. upstream retired
  a badge service or fixed a broken badge URL) that needs to be re-adapted with our own name,
  not dropped.
- **This fork has features upstream doesn't**: fake DNS answered natively in-tunnel
  (`dns/fakedns.go`, `dns/server.go`, `tunnel/tcp.go`/`tunnel/udp.go`), DNS hijacking as a
  port-53 catch-all (`dns/hijack.go`), configurable TCP keepalive (`core/option/option.go`,
  `engine/key.go`, `engine/engine.go`), and an Android AAR build/publish step in
  `.github/workflows/release.yml` (JDK + NDK + `gomobile bind`). Any upstream change touching
  the same files needs manual reconciliation, not a mechanical cherry-pick.
- **This fork already fixed some things upstream later fixed too** (e.g. the `windows/arm32`
  target being unbuildable on current Go toolchains). Expect no-ops; don't reapply something
  that's already identical.

## Procedure

1. **Fetch upstream, don't assume a local clone exists.**
   ```
   git remote add upstream https://github.com/xjasonlyu/tun2socks.git   # if not already present
   git fetch upstream --tags
   ```
   Never fetch from a local checkout path — always the real GitHub repo, so this works
   regardless of what other clones happen to exist on the machine.

2. **Resolve the target tag.** If the user gave a version without a `v` prefix or an
   ambiguous ref, resolve it against `git tag --list` on the `upstream` remote
   (`git ls-remote --tags upstream` if not yet fetched). Confirm the resolved tag with the user
   only if it's genuinely ambiguous — otherwise proceed.

3. **Find the fork's actual divergence point**, not just "upstream/main":
   ```
   git merge-base main upstream/<last-synced-tag-or-main>
   ```
   If unsure what the last sync point was, use `git merge-base main <target-tag>` directly —
   this finds the most recent commit both histories share, which is what actually matters.
   Then list what's new:
   ```
   git log --oneline <merge-base>..<target-tag>
   git diff --stat <merge-base>..<target-tag>
   ```

4. **If the diffstat is empty or touches nothing but files this fork doesn't meaningfully
   track changes in, say so and stop** — don't create a branch or commit for a no-op sync.
   This has happened before (v2.7.0 was CI/chore-only for this fork).

5. **Triage each upstream commit in the range**:
   - Pure branding/identity revert (e.g. `xjasonlyu` shows up again in a file this fork
     renamed) → skip the revert, but re-derive the *substantive* part of the change (if any)
     against this fork's own name. Example precedent: upstream retired a Go Report Card badge
     and replaced it with Go Build/Go Linter badges — that's real content, so it was
     re-applied against `Stanislav-Povolotsky/tun2socks` links, not skipped outright.
   - CI workflow changes (action version bumps, permission blocks, new steps) → apply to this
     fork's own `.github/workflows/*.yml`, preserving any fork-only steps (the Android AAR
     build in `release.yml` especially) that upstream's version of the file doesn't have.
   - Engine/core/proxy/tunnel code changes → apply directly if the surrounding code hasn't
     diverged, or hand-merge if it has (check `dns/`, `tunnel/tcp.go`, `tunnel/udp.go`,
     `engine/engine.go`, `engine/key.go` especially — these carry this fork's own features).
   - Dependency bumps (`go.mod`/`go.sum`) → apply, then run `go build ./... && go vet ./... &&
     go test ./...` to confirm nothing broke.

6. **Work on a new branch off current `main`**, never commit directly to `main`:
   ```
   git checkout -b merge-upstream-<target-tag> main
   ```

7. **Verify before committing**: `go build ./...`, `go vet ./...`, `go test ./...` must all
   pass.

8. **Commit with a message that names what was actually taken, what was skipped and why, and
   what was adapted** — not just "merge upstream vX.Y.Z". Future readers (including a future
   run of this same skill) need to know the reasoning, not just the diff. Splitting into
   multiple focused commits (e.g. one for CI bumps, one for a real feature port) is fine and
   often clearer than one giant commit — match the granularity to how unrelated the changes
   actually are.

9. **Don't push.** Report the branch name and a summary of what changed; the user pushes and
   merges themselves (established workflow for this repo).

## Reference: last known-good sync

- Fork diverged from upstream at `dda1b10` ("Chore: bump go mods (#537)").
- Synced through upstream `v2.7.0` (`8dda19e`) on branch `merge-upstream-v2.7.0`: CI action
  version bumps + README badge fix carried over; `windows/arm32` removal was already done
  independently (no-op).
