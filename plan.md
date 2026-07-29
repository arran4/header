1. **Update `types.go`**:
   - Add App Links properties (`AppLinksIOSName`, `AppLinksIOSAppStoreID`, `AppLinksIOSURL`, `AppLinksAndroidName`, `AppLinksAndroidPackage`, `AppLinksAndroidURL`).
   - Add Instagram specific properties (`InstagramOwnerUserID`).
   - Add ApplicationTitle (`ApplicationTitle`) and ThemeColor (`ThemeColor`).
   - Update `GroupedProperties` to include these new fields.

2. **Update `opengraph.go`**:
   - In the `meta` tag parser, extract `al:ios:*`, `al:android:*`, `instapp:owner_user_id`, `theme-color`, and `application-title`.

3. **Update Linter Rules**:
   - Create a `discord` rule in `lint/rules/discord/discord.go`.
   - The rule will check if the intent is to show video (like URL contains `/reels/` or `/shorts/` or `/reel/` or `og:video` is present) and if it is, verify that `twitter:player` and `oembed` link (or just `twitter:player`) is also present. If missing, it should report that it's missing what Discord requires.
   - Register the `discord` rule in `lint/lists/all/all.go` and `lint/lists/recommended/recommended.go`.

4. **Verify Discord Rule**:
   - Use `read_file` to verify the correct creation of `lint/rules/discord/discord.go` and the updates to `lint/lists/all/all.go` and `lint/lists/recommended/recommended.go`.

5. **Update tests and expectations**:
   - Create `testdata/extractor/instagram.txtar` with the mock HTML and expected JSON to test the new fields.
   - Update the expected outputs in `testdata/cli/txtar/lint_lists.txtar`, `testdata/cli/txtar/lint_list_rules.txtar`, `testdata/cli/txtar/lint_root_cmds.txtar`, and `testdata/cli/txtar/cli.txtar` (if any are affected) to include the newly registered `discord` rule and the new extracted fields.

6. **Run Tests**:
   - Run `go test ./...` to ensure all tests pass and no regressions were introduced.

7. **Pre-commit Instructions**:
   - Complete pre-commit steps to ensure proper testing, verification, review, and reflection are done.
