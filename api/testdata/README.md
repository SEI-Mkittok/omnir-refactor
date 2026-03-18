# Test Data Fixtures

This directory contains test fixtures for Phase 6 QA testing.

## CSV Import Fixtures

- **import_contacts_valid.csv** - 10 valid contact rows with all fields
- **import_contacts_malformed.csv** - Malformed CSV (missing closing quote) for error handling tests
- **import_accounts_valid.csv** - 5 valid account rows

## Generating Large Test Files

To generate a large CSV file for performance testing:

```bash
cd api/testdata
go run ../scripts/generate_large_csv.go -rows 10000 -output import_large.csv
```

## Email Fixtures (TODO)

- **email_plain.eml** - Plain text email sample
- **email_html.eml** - HTML email with attachment
- **email_thread.eml** - Reply email with In-Reply-To header

## Usage in Tests

```go
import "os"

func loadCSV(t *testing.T, filename string) []byte {
    data, err := os.ReadFile(filepath.Join("testdata", filename))
    require.NoError(t, err)
    return data
}
```
