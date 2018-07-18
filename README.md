# regression-gitbase

**regression-gitbase** is a tool than runs different versions of `gitbase` and compares its resource consumption.

```
Usage:
  regression [OPTIONS]

gitbase regression tester.

This tool executes several versions of gitbase and compares query times and
resource usage. There should be at least two versions specified as arguments
in the following way:

* v0.12.1 - release name from github
(https://github.com/src-d/gitbase/releases). The binary will be downloaded.
* latest - latest release from github. The binary will be downloaded.
* remote:master - any tag or branch from gitbase repository. The binary will
be built automatically.
* local:fix/some-bug - tag or branch from the repository in the current
directory. The binary will be built.
* pull:266 - code from pull request #266 from gitbase repo. Binary is built.
* /path/to/gitbase - a gitbase binary built locally.

The repositories and downloaded/built gitbase binaries are cached by default
in "repos" and "binaries" repositories from the current directory.


Application Options:
      --binaries=   Directory to store binaries (default: binaries)
                    [$REG_BINARIES]
      --repos=      Directory to store repositories (default: repos)
                    [$REG_REPOS]
      --url=        URL to the tool repo [$REG_GITURL]
      --gitport=    Port for local git server (default: 9418) [$REG_GITPORT]
  -c, --complexity= Complexity of the repositories to test (default: 1)
                    [$REG_COMPLEXITY]
  -n, --repeat=     Number of times a test is run (default: 3) [$REG_REPEAT]
      --show-repos  List available repositories to test

Help Options:
  -h, --help        Show this help message
```

To run this, you will need to have a working installation of [pilosa](https://github.com/pilosa/pilosa).

If you're manually running the regression-gitbase binary, you just need to have a pilosa server available. A pilosa server can be started using docker with the following command:

```
docker run -d --name pilosa -p 127.0.0.1:10101:10101 pilosa/pilosa:v0.9.0
```

## License

Licensed under the terms of the Apache License Version 2.0. See the `LICENSE`
file for the full license text.

