package mysheetsql

import (
	"fmt"
	"io"

	"github.com/src-d/go-mysql-server/sql"
	"google.golang.org/api/sheets/v4"
	"gopkg.in/src-d/go-log.v1"
)

// requestSize is the number of rows per API request
const requestSize = 20

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

		valuesFunc := func(from, to uint64) (*sheets.ValueRange, error) {
			readRange := fmt.Sprintf("%v!%v:%v", sheetName, from, to)
			resp, err := db.srv.Spreadsheets.Values.Get(db.spreadsheetId, readRange).Do()
			if err != nil {
				return nil, fmt.Errorf("unable to retrieve data from sheet range (%v): %v", readRange, err)
			}
			return resp, err
		}

		tables[sheetName] = &Table{
			values:    valuesFunc,
			sheetName: sheetName,
			schema:    schema,
		}
	}

	return tables
}

// Table is an individual sheet inside the document
type Table struct {
	values func(from, to uint64) (*sheets.ValueRange, error)

	sheetName string
	schema    sql.Schema
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
		values:   t.values,
		sheetRow: 2, // 1st row is the column names, values start on 2
		width:    len(t.schema),
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
	values func(from, to uint64) (*sheets.ValueRange, error)

	// iterator index over the sheet row, starts at 1
	sheetRow uint64
	// width of the rows
	width int
	// current index over sheetValues
	pos uint64
	// values fetched from the API. Each new request replaces the previous values
	sheetValues [][]interface{}
}

func (i *rowIter) Next() (sql.Row, error) {
	if i.pos >= uint64(len(i.sheetValues)) {
		err := i.nextValues()
		if err != nil {
			return nil, err
		}
	}

	row := make([]interface{}, i.width)
	allEmpty := true
	for j := range row {
		if len(i.sheetValues[i.pos]) <= j {
			break
		}

		v := i.sheetValues[i.pos][j]
		row[j] = v
		if v.(string) != "" {
			allEmpty = false
		}
	}

	if allEmpty {
		return nil, io.EOF
	}

	i.pos++

	return row, nil
}

// nextValues calls the API to retrieve the next requestSize rows. It also resets
// i.pos to 0
func (i *rowIter) nextValues() error {
	resp, err := i.values(i.sheetRow, i.sheetRow+requestSize)
	if err != nil {
		return err
	}

	if len(resp.Values) == 0 {
		return io.EOF
	}

	i.sheetValues = resp.Values
	i.sheetRow += uint64(len(resp.Values))

	i.pos = 0

	return nil
}

func (i *rowIter) Close() error {
	return nil
}
