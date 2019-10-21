# mysql-sheets
Allows Google Sheets access from a MySQL client

**Work in progress**

To use:

- Download `credentials.json` from https://developers.google.com/sheets/api/quickstart/go#step_1_turn_on_the
- `go run cmd/mysheetsql/main.go`
- `mysql -u user -h 127.0.0.1 -P 3306 -ppass`
