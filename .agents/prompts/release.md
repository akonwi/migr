---
description: Create a migr release: write notes, build, tag, and publish
---
Create a new migr release.

## Instructions

1. **Determine the version**. If no version argument was given, scan commits since the last tag to determine the bump:
   - `git log $(git describe --tags --abbrev=0 2>/dev/null || echo HEAD)..HEAD --oneline`
   - **Major** — any commit message contains `BREAKING CHANGE` or a `!` after the type (e.g. `feat!:`)
   - **Minor** — any commit matches `feat:` or `feature:`
   - **Patch** — otherwise (fixes, chores, docs, refactors, etc.)
   - If no tags exist yet, propose `v0.1.0`
   - Suggest the version and ask for confirmation.

2. **Generate release notes**. Gather notable changes since the last tag:
   - Scan commits with `git log` between the last tag and HEAD
   - Group by type: New Features, Improvements, Bug Fixes, Chores
   - Keep them concise and user-facing

3. **Run the release script** with `--dry-run` first:
   ```bash
   bash scripts/release.sh --dry-run v{version}
   ```
   Verify it succeeds and show the output summary.

4. **Ask for confirmation** to proceed with:
   - Running the script for real (creates the tag)
   - Pushing the tag to GitHub
   - Creating the GitHub Release with the generated notes
   - Updating the formula in `../homebrew-tap/Formula/migr.rb`

5. **If confirmed proceed step by step**:
   a. Run `bash scripts/release.sh v{version}`
   b. Run `git push origin v{version}`
   c. Publish: `cd ard-out/release && gh release create v{version} *.tar.gz --title "v{version}" --notes "..."` (pass the generated notes)
   d. Update `../homebrew-tap/Formula/migr.rb` with the new version and SHA256s (take them from the script output), then commit and push
