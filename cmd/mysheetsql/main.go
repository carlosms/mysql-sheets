package main

import (
	"io/ioutil"
	"log"

	mysheetsql "github.com/carlosms/mysql-sheets"

	sqle "github.com/src-d/go-mysql-server"
	"github.com/src-d/go-mysql-server/auth"
	"github.com/src-d/go-mysql-server/server"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/sheets/v4"
)

func main() {
	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets.readonly")
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config. See https://developers.google.com/sheets/api/quickstart/go#step_1_turn_on_the: %v", err)
	}
	client := mysheetsql.GetClient(config)

	srv, err := sheets.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}

	const spreadsheetId = "1Mal5p_TADNL6N_rfBAZatbShphyb3j8Bhu3ATD128RE"

	driver := sqle.NewDefault()
	driver.AddDatabase(mysheetsql.NewDatabase(spreadsheetId, srv))

	sqlConfig := server.Config{
		Protocol: "tcp",
		Address:  "localhost:3306",
		Auth:     auth.NewNativeSingle("user", "pass", auth.AllPermissions),
	}

	s, err := server.NewDefaultServer(sqlConfig, driver)
	if err != nil {
		log.Fatalf(err.Error())
	}

	s.Start()
}
