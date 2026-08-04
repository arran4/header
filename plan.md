1. **Update `types.go`**: Add missing JSON-LD fields to the `PageInfo` struct: `JSONLDAuthor`, `JSONLDDuration`, `JSONLDUploadDate`, `JSONLDDescription`, `JSONLDThumbnailURL`, `JSONLDID`.
   Update the `GroupedProperties` method to include these new fields in the output.
2. **Update `jsonld.go`**: Update `extractDataFromMap` to extract the new fields:
    - Author (can be a string or an object with a "name" field)
    - Duration (can check for "duration")
    - UploadDate ("uploadDate")
    - Description ("description")
    - ThumbnailUrl ("thumbnailUrl")
    - ID ("@id")
3. **Update tests**:
   - Update `jsonld_test.go` and `testdata/jsonld.html` to cover the new fields.
   - Update any `txtar` expected test files that rely on JSON output (e.g., in `templates/testdata/` or `testdata/extractor/`).
4. **Pre-commit**: Complete the pre-commit instructions required for the project.
5. **Submit**: Once verified, submit changes.
