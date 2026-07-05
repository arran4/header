
## Testing Output Changes
Any time you modify how output is displayed, you MUST update the `txtar` tests (like `testdata/cli.txt`) to reflect the new output format, ensuring you run `go test ./...` and verify that the scripts pass. This helps ensure changes in formatting are explicitly reviewed and tested.
