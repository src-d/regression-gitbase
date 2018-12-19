package gitbase

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"gopkg.in/src-d/go-log.v1"
	"gopkg.in/src-d/regression-core.v0"
)

type (
	gitbaseResults map[string][]*Result
	versionResults map[string]gitbaseResults

	// Test holds the information about a gitbase test.
	Test struct {
		config    regression.Config
		repos     *regression.Repositories
		testRepos string
		gitbase   map[string]*regression.Binary
		results   versionResults
		queries   []Query
		log       log.Logger
	}
)

// NewTest creates a new Test struct.
func NewTest(config regression.Config) (*Test, error) {
	repos, err := regression.NewRepositories(config)
	if err != nil {
		return nil, err
	}

	l, err := (&log.LoggerFactory{Level: log.InfoLevel}).New(log.Fields{})
	if err != nil {
		return nil, err
	}

	return &Test{
		config:  config,
		repos:   repos,
		queries: nil,
		log:     l,
	}, nil
}

// Prepare downloads repos and binaries needed for the test.
func (t *Test) Prepare() error {
	err := t.prepareRepos()
	if err != nil {
		return err
	}

	err = t.prepareGitbase()
	return err
}

// Run executes the tests.
func (t *Test) Run() error {
	results := make(versionResults)

	for _, version := range t.config.Versions {
		_, ok := results[version]
		if !ok {
			results[version] = make(gitbaseResults)
		}

		gitbase, ok := t.gitbase[version]
		if !ok {
			panic("gitbase not initialized. Was Prepare called?")
		}

		l := t.log.New(log.Fields{"version": version})

		l.Infof("Running version tests")

		times := t.config.Repeat
		if times < 1 {
			times = 1
		}

		rf := gitbase.ExtraFile("regression.yml")
		if queries, err := loadQueriesYaml(rf); err != nil {
			return err
		} else {
			t.queries = queries
		}

		for _, query := range t.queries {
			results[version][query.ID] = make([]*Result, times)

			for i := 0; i < times; i++ {
				l.New(log.Fields{
					"query.ID":   query.ID,
					"query.Name": query.Name,
				}).Infof("Running query")

				result, err := t.runTest(gitbase, t.testRepos, query)
				results[version][query.ID][i] = result

				// TODO: do not stop on errors ???
				if err != nil {
					return err
				}
			}
		}
	}

	t.results = results

	return nil
}

func (t *Test) PrintTabbedResults() {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 0, ' ', tabwriter.TabIndent|tabwriter.Debug)
	fmt.Fprint(w, "\x1b[1;33m ID \x1b[0m")
	versions := t.config.Versions
	for _, v := range versions {
		fmt.Fprintf(w, "\t\x1b[1;33m %s \x1b[0m", v)
	}
	fmt.Fprintf(w, "\n")

	for _, q := range t.queries {
		fmt.Fprintf(w, "\x1b[1;37m %s \x1b[0m", q.ID)
		var (
			mini    int
			min     time.Duration
			maxi    int
			max     time.Duration
			results []string
		)
		for i, v := range versions {
			if r, found := t.results[v][q.ID]; !found {
				results = append(results, "--")
			} else {
				t := r[0].Wtime
				for _, ri := range r[1:] {
					if ri.Wtime < min {
						t = ri.Wtime
					}
				}

				if min == 0 {
					min = t
				}

				if max == 0 {
					max = t
				}

				if t < min {
					min = t
					mini = i
				}

				if t > max {
					max = t
					maxi = i
				}

				results = append(results, t.String())
			}
		}

		for i, r := range results {
			fmt.Fprint(w, "\t")
			if i == mini {
				fmt.Fprintf(w, "\x1b[1;32m %s \x1b[0m", r)
			} else if i == maxi {
				fmt.Fprintf(w, "\x1b[1;31m %s \x1b[0m", r)
			} else {
				fmt.Fprintf(w, "\x1b[1;37m %s \x1b[0m", r)
			}
		}
		fmt.Fprintf(w, "\n")
	}
	w.Flush()
	fmt.Println()
}

func average(pr []*Result) *regression.Result {
	if len(pr) == 0 {
		return nil
	}

	results := make([]*regression.Result, 0, len(pr))
	for _, r := range pr {
		results = append(results, r.Result)
	}

	return regression.Average(results)
}

func (t *Test) SaveLatestCSV() {
	version := t.config.Versions[len(t.config.Versions)-1]
	for _, q := range t.queries {
		res := average(t.results[version][q.ID])
		if err := res.SaveAllCSV(fmt.Sprintf("plot_%s_", q.ID)); err != nil {
			panic(err)
		}
	}
}

// GetResults prints test results and returns if the tests passed.
func (t *Test) GetResults() bool {
	if len(t.config.Versions) < 1 {
		panic("there should be at least one version")
	}

	versions := t.config.Versions
	ok := true
	for i, version := range versions[0 : len(versions)-1] {
		fmt.Printf("%s - %s ####\n", version, versions[i+1])
		a := t.results[versions[i]]
		b := t.results[versions[i+1]]

		for _, query := range t.queries {
			fmt.Printf("## Query {ID: %s, Name: %s} ##\n", query.ID, query.Name)
			if _, found := a[query.ID]; !found {
				fmt.Printf("# Skip - Query.ID: %s not found for version: %s\n", query.ID, versions[i])
				continue
			}
			if _, found := b[query.ID]; !found {
				fmt.Printf("# Skip - Query.ID: %s not found for version: %s\n", query.ID, versions[i+1])
				continue
			}

			queryA := a[query.ID][0]
			queryB := b[query.ID][0]

			queryA.Result = average(a[query.ID])
			queryB.Result = average(b[query.ID])
			c := queryA.ComparePrint(queryB, 10.0)
			if !c {
				ok = false
			}
		}
	}

	return ok
}

func (t *Test) runTest(
	gitbase *regression.Binary,
	repos string,
	query Query,
) (*Result, error) {
	t.log.Infof("Executing gitbase test")

	server := NewServer(gitbase.Path, repos)
	err := server.Start()
	if err != nil {
		t.log.With(log.Fields{
			"repos":   repos,
			"gitbase": gitbase.Path,
		}).Errorf(err, "Could not execute gitbase")
		return nil, err
	}

	queries := NewSQLTest(server.URL(), query)
	err = queries.Connect()
	if err != nil {
		return nil, err
	}

	start := time.Now()

	rows, err := queries.Execute()
	if err != nil {
		return nil, err
	}

	wall := time.Since(start)

	queries.Disconnect()
	server.Stop()

	rusage := server.Rusage()

	t.log.With(log.Fields{
		"wall":   wall,
		"memory": rusage.Maxrss,
	}).Infof("Finished queries")

	result := &regression.Result{
		Wtime:  wall,
		Stime:  time.Duration(rusage.Stime.Nano()),
		Utime:  time.Duration(rusage.Utime.Nano()),
		Memory: rusage.Maxrss * 1024,
	}

	r := &Result{
		Result: result,
		Query:  query,
		Rows:   rows,
	}

	return r, nil
}

func (t *Test) prepareRepos() error {
	t.log.Infof("Downloading repositories")
	err := t.repos.Download()
	if err != nil {
		return err
	}

	t.testRepos, err = t.repos.LinksDir()
	return err
}

func (t *Test) prepareGitbase() error {
	t.log.Infof("Preparing gitbase binaries")
	releases := regression.NewReleases("src-d", "gitbase", t.config.GitHubToken)

	t.gitbase = make(map[string]*regression.Binary, len(t.config.Versions))
	for _, version := range t.config.Versions {
		b := NewGitbase(t.config, version, releases)
		err := b.Download()
		if err != nil {
			return err
		}

		t.gitbase[version] = b
	}

	return nil
}
