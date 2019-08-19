package gitbase

import (
	"context"
	"database/sql"
	"io/ioutil"

	// Load mysql drivers.
	_ "github.com/go-sql-driver/mysql"
	"gopkg.in/yaml.v2"
)

// Query struct has information about on query. It can consist on more than
// one statement.
type Query struct {
	ID         string   `yaml:"ID"`
	Name       string   `yaml:"Name,omitempty"`
	Statements []string `yaml:"Statements"`
}

// SQLTest holds are the queries that belong to a test and connection
// functionality.
type SQLTest struct {
	Query Query
	URL   string
	db    *sql.DB
}

// NewSQLTest creates a new SQLTest.
func NewSQLTest(url string, query Query) *SQLTest {
	return &SQLTest{
		Query: query,
		URL:   url,
	}
}

// Connect creates the mysql connection to gitbase.
func (q *SQLTest) Connect() error {
	db, err := sql.Open("mysql", q.URL)
	if err != nil {
		return err
	}

	q.db = db

	return nil
}

// Disconnect closes the mysql connection.
func (q *SQLTest) Disconnect() error {
	return q.db.Close()
}

func (q *SQLTest) ExecuteCtx(ctx context.Context) (int64, error) {
	var count int64

	for _, s := range q.Query.Statements {
		rows, err := q.db.QueryContext(ctx, s)
		if err != nil {
			return 0, err
		}

		for rows.Next() {
			count++
		}
	}

	return count, nil
}

// Execute runs sql query on the gitbase server.
func (q *SQLTest) Execute() (int64, error) {
	var count int64

	for _, s := range q.Query.Statements {
		rows, err := q.db.Query(s)
		if err != nil {
			return 0, err
		}

		for rows.Next() {
			count++
		}
	}

	return count, nil
}

func loadQueriesYaml(file string) ([]Query, error) {
	text, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var q []Query
	err = yaml.Unmarshal(text, &q)
	if err != nil {
		return nil, err
	}

	return q, nil
}
