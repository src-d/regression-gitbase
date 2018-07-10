package gitbase

import (
	"database/sql"

	// Load mysql drivers.
	_ "github.com/go-sql-driver/mysql"
)

// DefaultQueries has a list of queries executed by the tool.
var DefaultQueries = []Query{
	{
		"all commits",
		[]string{
			"select * from commits",
		},
	},
	{
		"Last commit messages in HEAD for every repository",
		[]string{`
		SELECT c.commit_message
		FROM
			refs r
			JOIN commits c ON r.commit_hash = c.commit_hash
		WHERE
			r.ref_name = 'HEAD';`},
	},
	{
		"All commit messages in HEAD history for every repository",
		[]string{`
		SELECT c.commit_message
		FROM commits c
		NATURAL JOIN ref_commits r
		WHERE r.ref_name = 'HEAD';`},
	},
	{
		"Top 10 repositories by commit count in HEAD",
		[]string{`
		SELECT
			repository_id,
			commit_count
		FROM (
			SELECT
				r.repository_id,
				count(*) AS commit_count
			FROM ref_commits r
			WHERE r.ref_name = 'HEAD'
			GROUP BY r.repository_id
		) AS q
		ORDER BY commit_count DESC
		LIMIT 10;`},
	},
	// Disabled as language call makes the query super slow
	//
	// {
	// 	"Repository count by language presence (HEAD, no forks)",
	// 	[]string{`
	// 	SELECT *
	// 	FROM (
	// 		SELECT
	// 			language,
	// 			COUNT(repository_id) AS repository_count
	// 		FROM (
	// 		SELECT DISTINCT
	// 			r.repository_id AS repository_id,
	// 			language(f.file_path, f.blob_content) AS language
	// 		FROM refs r
	// 		NATURAL JOIN commit_trees
	// 		NATURAL JOIN files f
	// 		WHERE refs.ref_name = 'HEAD'
	// 		) AS q1
	// 		GROUP BY language
	// 	) AS q2
	// 	ORDER BY repository_count DESC;`},
	// },
	{
		"Top 10 repositories by contributor count (all branches)",
		[]string{`
		SELECT
			repository_id,
			contributor_count
		FROM (
			SELECT
				repository_id,
				COUNT(DISTINCT commit_author_email) AS contributor_count
			FROM commits
			GROUP BY repository_id
		) AS q
		ORDER BY contributor_count DESC
		LIMIT 10;`},
	},
	{
		"Create index on language UDF",
		[]string{`CREATE INDEX language_idx
		ON files(language(file_path, blob_content))
		WITH (async = false)`},
	},
	{
		"Query by language using the index",
		[]string{
			`CREATE INDEX language_idx
			ON files(language(file_path, blob_content))
			WITH (async = false)`,

			`SELECT file_path
			FROM files
			WHERE language(file_path, blob_content) = 'Go'`,

			`DROP INDEX language_idx ON files`,
		},
	},
	{
		"Query all files from HEAD",
		[]string{`
		SELECT cf.file_path, f.blob_content
		FROM ref_commits r
		NATURAL JOIN commit_files cf
		NATURAL JOIN files f
		WHERE r.ref_name = 'HEAD' AND r.index = 0`},
	},
	{
		"Get all LICENSE blobs using index",
		[]string{
			`CREATE INDEX file_path_idx
			ON files(file_path) WITH (async = false)`,

			`SELECT blob_content
			FROM files
			WHERE file_path = 'LICENSE'`,
		},
	},
}

// Query struct has information about on query. It can consist on more than
// one statement.
type Query struct {
	Name       string
	Statements []string
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
