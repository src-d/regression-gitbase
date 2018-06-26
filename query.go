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
		"all repositories",
		[]string{
			"select * from repositories",
		},
	},
	{
		"Count repositories",
		[]string{`
	SELECT COUNT(DISTINCT repository_id) AS repository_count
	FROM repositories`},
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
	// {
	// 	"All commit messages in HEAD history for every repository",
	// 	[]string{`
	// SELECT c.message
	// FROM
	// 	commits c
	// 	JOIN refs r ON r.commit_hash = c.commit_hash
	// WHERE
	// 	r.name = 'refs/heads/HEAD' AND
	// 	history_idx(r.hash, c.hash) >= 0;`},
	// },
	// {
	// 	"Top 10 repositories by commit count in HEAD",
	// 	[]string{`
	// SELECT
	// 	repository_id,
	// 	commit_count
	// FROM (
	// 	SELECT
	// 		r.repository_id,
	// 		count(*) AS commit_count
	// 	FROM
	// 		refs r
	// 		JOIN commits c ON history_idx(r.hash, c.hash) >= 0
	// 	WHERE
	// 		r.name = 'refs/heads/HEAD'
	// 	GROUP BY r.repository_id
	// ) AS q
	// ORDER BY commit_count DESC
	// LIMIT 10;`},
	// },
	// {
	// 	"Count repository HEADs",
	// 	[]string{`
	// SELECT
	// 	COUNT(DISTINCT r.repository_id) AS head_count
	// FROM
	// 	refs r
	// WHERE name = 'refs/heads/HEAD';`},
	// },

	// {
	// 	"Repository count by language presence (HEAD, no forks)",
	// 	[]string{`
	// SELECT *
	// FROM (
	// 	SELECT
	// 		language,
	// 		COUNT(repository_id) AS repository_count
	// 	FROM (
	// 		SELECT DISTINCT
	// 			r.repository_id AS repository_id,
	// 			language(t.name, b.content) AS language
	// 		FROM
	// 			refs r
	// 			JOIN commits c ON r.hash = c.hash
	// 			JOIN tree_entries t ON commit_has_tree(c.hash, t.tree_hash)
	// 			JOIN blobs b ON t.entry_hash = b.hash
	// 		WHERE
	// 			r.name = 'refs/heads/HEAD'
	// 	) AS q1
	// 	GROUP BY language
	// ) AS q2
	// ORDER BY repository_count DESC;`},
	// },
	// {
	// 	"Top 10 repositories by contributor count (all branches)",
	// 	[]string{`
	// SELECT
	// 	repository_id,
	// 	contributor_count
	// FROM (
	// 	SELECT
	// 		repository_id,
	// 		COUNT(DISTINCT c.author_email) AS contributor_count
	// 	FROM
	// 		refs r
	// 		JOIN commits c ON history_idx(r.hash, c.hash) >= 0
	// 	GROUP BY repository_id
	// ) AS q
	// ORDER BY contributor_count DESC
	// LIMIT 10;`},
	// },
	// {
	// 	"Created projects per year",
	// 	[]string{`
	// SELECT
	// 	year,
	// 	COUNT(DISTINCT hash) AS project_count
	// FROM (
	// 	SELECT
	// 		hash,
	// 		YEAR(author_date) AS year
	// 	FROM
	// 		refs r
	// 		JOIN commits c ON r.hash = c.hash
	// 	WHERE
	// 		r.name = 'refs/heads/HEAD'
	// ) AS q
	// GROUP BY year
	// ORDER BY year DESC;`},
	// },
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
