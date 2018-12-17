package gitbase

import (
	"database/sql"
	"io/ioutil"

	// Load mysql drivers.
	_ "github.com/go-sql-driver/mysql"
	yaml "gopkg.in/yaml.v2"
)

// DefaultQueries has a list of queries executed by the tool if we could not find a gitbase/_testdata/regression.yml file.
// ID per query matches IDs in gitbase/_testdata/regression.yml (for easier comparision)
var DefaultQueries = []Query{
	{
		ID:   "query0",
		Name: "All commits",
		Statements: []string{
			"SELECT * FROM commits",
		},
	},
	{
		ID:   "query1",
		Name: "Last commit messages in HEAD for every repository",
		Statements: []string{
			`SELECT c.commit_message FROM refs r JOIN commits c ON r.commit_hash = c.commit_hash WHERE r.ref_name = 'HEAD'`,
		},
	},
	{
		ID:   "query2",
		Name: "All commit messages in HEAD history for every repository",
		Statements: []string{
			`SELECT c.commit_message FROM commits c NATURAL JOIN ref_commits r WHERE r.ref_name = 'HEAD'`,
		},
	},
	{
		ID:   "query3",
		Name: "Top 10 repositories by commit count in HEAD",
		Statements: []string{
			`SELECT repository_id,commit_count FROM (SELECT r.repository_id,count(*) AS commit_count FROM ref_commits r WHERE r.ref_name = "HEAD" GROUP BY r.repository_id) AS q ORDER BY commit_count DESC LIMIT 10`,
		},
	},
	{
		ID:   "query4",
		Name: "Top 10 repositories by contributor count (all branches)",
		Statements: []string{
			`SELECT repository_id,contributor_count FROM (SELECT repository_id, COUNT(DISTINCT commit_author_email) AS contributor_count FROM commits GROUP BY repository_id) AS q ORDER BY contributor_count DESC LIMIT 10`,
		},
	},
	// {
	// 	ID:   "query5",
	// 	Name: "Create pilosa index on language UDF",
	// 	Statements: []string{
	// 		`CREATE INDEX language_idx ON files USING pilosa (language(file_path, blob_content)) WITH (async = false)`,
	// 	},
	// },
	// {
	// 	ID:   "query7",
	// 	Name: "Query by language using the pilosa index",
	// 	Statements: []string{
	// 		`CREATE INDEX language_idx ON files USING pilosa (language(file_path, blob_content)) WITH (async = false)`,
	// 		`SELECT file_path FROM files WHERE language(file_path, blob_content) = 'Go'`,
	// 		`DROP INDEX language_idx ON files`,
	// 	},
	// },
	// {
	// 	ID:   "query10",
	// 	Name: "Get all LICENSE blobs using pilosa index",
	// 	Statements: []string{
	// 		`CREATE INDEX file_path_idx ON files USING pilosa (file_path) WITH (async = false)`,
	// 		`SELECT blob_content FROM files WHERE file_path = 'LICENSE'`,
	// 	},
	// },
	{
		ID:   "query12",
		Name: "10 top repos by file count in HEAD",
		Statements: []string{
			`SELECT repository_id, num_files FROM (SELECT COUNT(f.*) num_files, f.repository_id FROM ref_commits r INNER JOIN commit_files cf ON r.commit_hash = cf.commit_hash AND r.repository_id = cf.repository_id INNER JOIN files f ON cf.repository_id = f.repository_id AND cf.blob_hash = f.blob_hash AND cf.tree_hash = f.tree_hash AND cf.file_path = f.file_path WHERE r.ref_name = 'HEAD' GROUP BY f.repository_id) t ORDER BY num_files DESC LIMIT 10`,
		},
	},
	{
		ID:   "query13",
		Name: "Top committers per repository",
		Statements: []string{
			`SELECT * FROM (SELECT commit_author_email as author,repository_id as id,count(*) as num_commits FROM commits GROUP BY commit_author_email, repository_id) t ORDER BY num_commits DESC`,
		},
	},
	{
		ID:   "query14",
		Name: "Top committers in all repositories",
		Statements: []string{
			`SELECT * FROM (SELECT commit_author_email as author,count(*) as num_commits FROM commits GROUP BY commit_author_email) t ORDER BY num_commits DESC`,
		},
	},
	// {
	// 	ID:   "query15",
	// 	Name: "Union operation with pilosa index",
	// 	Statements: []string{
	// 		`CREATE INDEX file_path_idx ON files USING pilosa (file_path) WITH (async = false)`,
	// 		`SELECT blob_content FROM files WHERE file_path = 'LICENSE' OR file_path = 'README.md'`,
	// 		`DROP INDEX file_path_idx ON files`,
	// 	},
	// },
	{
		ID:   "query17",
		Name: "Count all commits with NOT operation",
		Statements: []string{
			`SELECT COUNT(*) FROM commits WHERE NOT(commit_author_email = 'non existing email address')`,
		},
	},
}

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
