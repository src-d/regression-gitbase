package main

import (
	"os"

	gitbase "github.com/src-d/regression-gitbase"

	flags "github.com/jessevdk/go-flags"
	"github.com/src-d/regression-core"
	"gopkg.in/src-d/go-log.v1"
)

var description = `gitbase regression tester.

This tool executes several versions of gitbase and compares query times and resource usage. There should be at least two versions specified as arguments in the following way:

* v0.12.1 - release name from github (https://github.com/src-d/gitbase/releases). The binary will be downloaded.
* latest - latest release from github. The binary will be downloaded.
* remote:master - any tag or branch from gitbase repository. The binary will be built automatically.
* local:fix/some-bug - tag or branch from the repository in the current directory. The binary will be built.
* local:HEAD - current state of the repository. Binary is built.
* pull:266 - code from pull request #266 from gitbase repo. Binary is built.
* /path/to/gitbase - a gitbase binary built locally.

The repositories and downloaded/built gitbase binaries are cached by default in "repos" and "binaries" repositories from the current directory.
`

type Options struct {
	regression.Config
	GitServerConfig regression.GitServerConfig

	CSV bool `long:"csv" description:"save csv files with last result"`

	// prometheus pushgateway related options
	Prometheus bool `long:"prom" description:"store latest results to prometheus"`
	PromConfig regression.PromConfig
	CIConfig   regression.CIConfig
}

func main() {
	options := Options{
		Config: regression.NewConfig(),
	}

	parser := flags.NewParser(&options, flags.Default)
	parser.LongDescription = description

	args, err := parser.Parse()
	if err != nil {
		if err, ok := err.(*flags.Error); ok {
			if err.Type == flags.ErrHelp {
				os.Exit(0)
			}
		}

		log.Errorf(err, "Could not parse arguments")
		os.Exit(1)
	}

	config := options.Config
	gitServerConfig := options.GitServerConfig
	if config.ShowRepos {
		repos, err := regression.NewRepositories(gitServerConfig)
		if err != nil {
			log.Errorf(err, "Could not get repositories")
			os.Exit(1)
		}

		repos.ShowRepos()
		os.Exit(0)
	}

	if len(args) < 1 {
		log.Errorf(nil, "There should be at least one version")
		os.Exit(1)
	}

	config.Versions = args

	test, err := gitbase.NewTest(config, gitServerConfig)
	if err != nil {
		panic(err)
	}

	log.Infof("Preparing run")
	err = test.Prepare()
	if err != nil {
		log.Errorf(err, "Could not prepare environment")
		os.Exit(1)
	}

	err = test.RunLoad()
	if err != nil {
		panic(err)
	}

	test.PrintTabbedResults()
	res := test.GetResults()
	if !res {
		os.Exit(1)
	}
	if options.CSV {
		test.SaveLatestCSV()
	}
	if options.Prometheus {
		if err := test.StoreLatestToPrometheus(options.PromConfig, options.CIConfig); err != nil {
			log.Errorf(err, "Could not store results to prometheus")
			os.Exit(1)
		}
	}
}
