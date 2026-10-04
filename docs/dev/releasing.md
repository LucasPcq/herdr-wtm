# Releasing

A release is a `v<version>` tag on `main` whose version matches `herdr-plugin.toml`. `scripts/install.sh` (herdr's `[[build]]` step) downloads exactly that version's archive, so the two must agree.

1. Move the `[Unreleased]` entries of `CHANGELOG.md` under `## [x.y.z] - YYYY-MM-DD` (see [changelog.md](changelog.md)), add the link reference at the bottom, and check the section reads well as release notes: `make release-notes VERSION=x.y.z`.
2. Set `version = "x.y.z"` in `herdr-plugin.toml`.
3. Commit both on `main` (`chore: release vx.y.z`), then tag and push:

   ```bash
   git tag vx.y.z && git push origin main vx.y.z
   ```

The Release workflow runs the tests, refuses a tag that disagrees with the manifest, extracts the release notes from `CHANGELOG.md`, and publishes the binaries with GoReleaser (darwin/linux × amd64/arm64, `checksums.txt`).

## Betas

A tag with a prerelease part (`v0.3.0-beta.1`) becomes a GitHub pre-release, with the `[Unreleased]` section as its notes; the manifest must carry the same version (`0.3.0-beta.1`). Testers install it with `herdr plugin install LucasPcq/herdr-wtm --ref v0.3.0-beta.1`.

## Checking a release config change

`make snapshot` builds every archive without publishing (needs `goreleaser`); CI runs the same on every pull request.
