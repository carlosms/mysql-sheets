package main

import (
	"fmt"
	"io/ioutil"

	mysheetsql "github.com/carlosms/mysql-sheets"

	sqle "github.com/src-d/go-mysql-server"
	"github.com/src-d/go-mysql-server/auth"
	"github.com/src-d/go-mysql-server/server"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/sheets/v4"
	cli "gopkg.in/src-d/go-cli.v0"
)

// Replaced during release build
var (
	version = "dev"
	build   = "dev"
)

var app = cli.New("mysheetsql", version, build, "MySQL server that reads from Google Sheets data")

type serveCommand struct {
	cli.Command `name:"serve" short-description:"starts the server" long-description:"starts the server"`

	SpreadsheetId string `short:"i" long:"id" env:"MYSHEETSQL_ID" required:"true" description:"Spreadsheet identifier from its URL, e.g. 1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms"`
}

var sub = app.AddCommand(&serveCommand{})

func main() {
	app.RunMain()
}

func (c *serveCommand) Execute(args []string) error {
	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		return fmt.Errorf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets.readonly")
	if err != nil {
		return fmt.Errorf("Unable to parse client secret file to config. See https://developers.google.com/sheets/api/quickstart/go#step_1_turn_on_the: %v", err)
	}
	client := mysheetsql.GetClient(config)

	srv, err := sheets.New(client)
	if err != nil {
		return fmt.Errorf("Unable to retrieve Sheets client: %v", err)
	}

	driver := sqle.NewDefault()
	driver.AddDatabase(mysheetsql.NewDatabase(c.SpreadsheetId, srv))

	sqlConfig := server.Config{
		Protocol: "tcp",
		Address:  "localhost:3306",
		Auth:     auth.NewNativeSingle("user", "pass", auth.AllPermissions),
	}

	s, err := server.NewDefaultServer(sqlConfig, driver)
	if err != nil {
		return fmt.Errorf("Failed to create a go-mysql-server Server: %v", err)
	}

	s.Start()

	return nil
}
