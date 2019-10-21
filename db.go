package mysheetsql

import (
	"fmt"
	"io"

	"github.com/src-d/go-mysql-server/sql"
	"google.golang.org/api/sheets/v4"
	"gopkg.in/src-d/go-log.v1"
)

// Database is a Google Sheets document
type Database struct {
	spreadsheetId string
	srv           *sheets.Service
}

// NewDatabase creates a new DB for the given Google Sheets Id.
func NewDatabase(spreadsheetId string, srv *sheets.Service) Database {
	return Database{spreadsheetId: spreadsheetId, srv: srv}
}

// Name returns the name
func (db Database) Name() string {
	return db.spreadsheetId
}

// Tables returns the information of all tables.
func (db Database) Tables() map[string]sql.Table {
	spreadsheet, err := db.srv.Spreadsheets.Get(db.spreadsheetId).Do()
	if err != nil {
		log.Errorf(err, "unable to retrieve spreadsheet info")
		return nil
	}

	tables := make(map[string]sql.Table)

	for _, sheet := range spreadsheet.Sheets {
		sheetName := sheet.Properties.Title

		if sheet.Properties.SheetType != "GRID" {
			log.Infof("sheet %v skipped, it's not a grid type", sheetName)
			continue
		}

		if sheet.Properties.GridProperties.FrozenRowCount == 0 {
			log.Infof("sheet %v skipped, it needs to have 1 frozen row header with column names", sheetName)
			continue
		}

		schema := []*sql.Column{}

		// Get the header row with column names
		readRange := fmt.Sprintf("%v!1:1", sheetName)
		resp, err := db.srv.Spreadsheets.Values.Get(db.spreadsheetId, readRange).Do()
		if err != nil {
			log.Errorf(err, "unable to retrieve column names from sheet")
			continue
		}
		if len(resp.Values) == 0 {
			log.Errorf(nil, "unable to retrieve column names from sheet")
			continue
		}

		row := resp.Values[0]
		for _, col := range row {
			schema = append(schema,
				&sql.Column{
					Name:     col.(string),
					Nullable: true,
					Source:   sheetName,
					Type:     sql.Text,
				})
		}

		tables[sheetName] = &Table{
			spreadsheetId: db.spreadsheetId,
			sheetName:     sheetName,
			schema:        schema,
			srv:           db.srv,
		}
	}

	return tables
}

// Table is an individual sheet inside the document
type Table struct {
	spreadsheetId string
	sheetName     string
	schema        sql.Schema
	srv           *sheets.Service
}

// Name returns the name
func (t *Table) Name() string {
	return t.sheetName
}

func (t *Table) String() string {
	return fmt.Sprintf("Table %s", t.sheetName)
}

func (t *Table) Schema() sql.Schema {
	return t.schema
}

func (t *Table) Partitions(ctx *sql.Context) (sql.PartitionIter, error) {
	return &partitionIter{}, nil
}

func (t *Table) PartitionRows(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	return &rowIter{
		spreadsheetId: t.spreadsheetId,
		sheetName:     t.sheetName,
		sheetRow:      2,
		srv:           t.srv,
		size:          len(t.schema),
	}, nil
}

type partitionIter struct {
	pos int
}

func (i *partitionIter) Close() error {
	return nil
}

func (i *partitionIter) Next() (sql.Partition, error) {
	if i.pos > 0 {
		return nil, io.EOF
	}
	i.pos++

	return partition{}, nil
}

type partition struct{}

func (p partition) Key() []byte {
	return []byte("0")
}

type rowIter struct {
	spreadsheetId string
	sheetName     string
	sheetRow      uint64
	srv           *sheets.Service
	size          int
}

func (i *rowIter) Next() (sql.Row, error) {
	readRange := fmt.Sprintf("%v!%v:%v", i.sheetName, i.sheetRow, i.sheetRow)
	resp, err := i.srv.Spreadsheets.Values.Get(i.spreadsheetId, readRange).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve data from sheet range (%v): %v", readRange, err)
	}

	if len(resp.Values) == 0 {
		return nil, io.EOF
	}

	row := make([]interface{}, i.size)
	allEmpty := true
	for i := range row {
		if len(resp.Values[0]) <= i {
			break
		}

		v := resp.Values[0][i]
		row[i] = v
		if v.(string) != "" {
			allEmpty = false
		}
	}

	if allEmpty {
		return nil, io.EOF
	}

	i.sheetRow++

	return row, nil
}

func (i *rowIter) Close() error {
	return nil
}
