
## Testing Output Changes
Any time you modify how output is displayed, you MUST update the `txtar` tests (like `testdata/cli.txt`) to reflect the new output format, ensuring you run `go test ./...` and verify that the scripts pass. This helps ensure changes in formatting are explicitly reviewed and tested.

Make sure to test both `-style default` and `-style plain` outputs whenever you are modifying templates or what is displayed in the standard CLI output.

Also be sure to cover edge cases:
- When a field differs from its defaults/request (e.g. Canonical Link vs URL).
- When a field matches its defaults/request.
- When a field is entirely absent.
