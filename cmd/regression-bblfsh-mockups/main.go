package main

import (
	"bytes"
	"context"
	"os"
	"strconv"
	"strings"
	"time"

	v2 "github.com/bblfsh/sdk/v3/protocol"
	"github.com/bblfsh/sdk/v3/uast/nodes"
	"github.com/bblfsh/sdk/v3/uast/nodes/nodesproto"
	"github.com/jessevdk/go-flags"
	"github.com/src-d/regression-core"
	gitbase "github.com/src-d/regression-gitbase"
	mockup "github.com/src-d/regression-gitbase/bblfsh-mockups"
	capture_output "github.com/src-d/regression-gitbase/capture-output"
	"gopkg.in/src-d/go-errors.v1"
	"gopkg.in/src-d/go-log.v1"
)

const (
	limit             = 10
	queryTimeoutTotal = 15 * time.Second

	captureOutputDelay = 500 * time.Millisecond
)

var (
	testErrParseErrorOnly    = []*v2.ParseError{{"only one parse error"}}
	testErrParseErrorSeveral = []*v2.ParseError{
		{"first parse error"},
		{"second parse error"},
		{"third parse error"}}
	testErrParseErrorNilSlice []*v2.ParseError = nil

	uastQuery = `select file_path, uast(blob_content) name 
from refs natural 
join commit_files natural 
join blobs 
where LANGUAGE(file_path) = 'Go' 
limit ` + strconv.Itoa(limit)
)

func main() {
	var testsFailed []string
	for _, t := range []struct {
		name string
		f    func() error
	}{
		{"TestResponseErrorWarnings", TestResponseErrorWarnings},
		{"TestBrokenUASTInResponseWarning", TestBrokenUASTInResponseWarning},
		{"TestParseErrorWarnings", TestParseErrorWarnings},
		{"TestQueryExecBeforeTimeout", TestQueryExecBeforeTimeout},
		// TODO(lwsanty): can possibly fail because of https://github.com/src-d/go-mysql-server/pull/801
		// TODO this test needs refactor after https://github.com/src-d/gitbase/issues/950
		{"TestQueryExecAfterTimeout", TestQueryExecAfterTimeout},
	} {
		t := t
		log.Infof("=====> %v", t.name)
		if err := t.f(); err != nil {
			log.Infof("error occurred: %v", err)
			testsFailed = append(testsFailed, t.name)
		}
		log.Infof("=====> done")
	}

	if len(testsFailed) > 0 {
		log.Infof("%v tests have failed:\n%+v", len(testsFailed), testsFailed)
		os.Exit(1)
	}

	log.Infof("ALL TESTS PASSED")
}

// TestQueryExecBeforeTimeout
// 1) prepare bblfsh mockup that performs sleep responseLag time during the ParseRequest handling
//    and returns specific errText error on parse request
// 2) run gitbase with GITBASE_CONNECTION_TIMEOUT = responseLag + 1 second
// 3) execute query uastQuery
// 4) check that no error was returned during the query execution
// 5) check gitbase stderr: there should be warnings with errText
// 6) the amount of warnings should be equal to limit
func TestQueryExecBeforeTimeout() error {
	const (
		errText = "parse error"

		responseLagSeconds       = 1
		responseLag              = responseLagSeconds * time.Second
		connectionTimeoutSeconds = responseLagSeconds + 1
	)

	return testWarnings(context.Background(), mockup.OptsV2{
		ParseResponse:    nil,
		ParseResponseLag: responseLag,
		ParseResponseErr: errors.NewKind(errText).New(),
	}, map[string]string{
		"GITBASE_CONNECTION_TIMEOUT": strconv.Itoa(connectionTimeoutSeconds),
	}, errTextUnableToParse(errTextRPC(errText)), limit)
}

// TestQueryExecAfterTimeout
// 1) prepare bblfsh mockup that performs sleep responseLag time during the ParseRequest
//	  and returns specific errText error on parse request
// 2) run gitbase with GITBASE_CONNECTION_TIMEOUT = responseLag - 1 second
// 3) execute query uastQuery
// 4) check that warning(or error?) related to timeout appeared?
func TestQueryExecAfterTimeout() error {
	const (
		errText             = "parse error"
		errTimeoutOnRowRead = "row read wait bigger than connection timeout"

		responseLagSeconds       = 2
		responseLag              = responseLagSeconds * time.Second
		connectionTimeoutSeconds = responseLagSeconds - 1
	)
	var errAssert = errors.NewKind("expected error containing %q, got: %v")

	err := testWarnings(context.Background(), mockup.OptsV2{
		ParseResponse:    nil,
		ParseResponseLag: responseLag,
		ParseResponseErr: errors.NewKind(errText).New(),
	}, map[string]string{
		"GITBASE_CONNECTION_TIMEOUT": strconv.Itoa(connectionTimeoutSeconds),
	}, "timeout", limit)
	if err == nil || !strings.Contains(err.Error(), errTimeoutOnRowRead) {
		return errAssert.New(errTimeoutOnRowRead, err)
	}

	return nil
}

// TestResponseErrorWarnings
// 1) prepare bblfsh mockup that returns specific errText error on parse request
// 2) run gitbase
// 3) execute query uastQuery
// 4) check that no error was returned during the query execution
// 5) check gitbase stderr: there should be warnings with errText
// 6) the amount of warnings should be equal to limit
func TestResponseErrorWarnings() error {
	const errText = "parse error"

	return testWarnings(context.Background(), mockup.OptsV2{
		ParseResponse:    nil,
		ParseResponseErr: errors.NewKind(errText).New(),
	}, nil, errTextUnableToParse(errTextRPC(errText)), limit)
}

// TestBrokenUASTInResponseWarning
// 1) prepare bblfsh mockup that returns broken UAST bytes sequence
// 2) run gitbase
// 3) execute query uastQuery
// 4) check that no error was returned during the query execution
// 5) check gitbase stderr: there should be warnings with "short read"
// 6) the amount of warnings should be equal to limit
func TestBrokenUASTInResponseWarning() error {
	return testWarnings(context.Background(), mockup.OptsV2{
		ParseResponse: &v2.ParseResponse{Language: "go", Uast: []byte{1}},
	}, nil, errTextUnableToParse("short read"), limit)
}

// TestParseErrorWarnings
// contains three subtests with 3 bblfsh mockups: without ParseErrors, with one ParseError and with three ParseErrors
// for each of the mockup:
// 1) run  gitbase
// 2) execute query uastQuery
// 3) check that no error was returned during the query execution
// 4) check that related warnings appeared
func TestParseErrorWarnings() error {
	bufData, err := getUASTBytes("a")
	if err != nil {
		return err
	}

	testWarningsParseErrors := func(parseErrs []*v2.ParseError) error {
		var (
			text    string
			repeats = limit
		)
		if len(parseErrs) == 0 {
			text = "level=warning"
			repeats = 1
		} else {
			var tmp []string
			for _, pe := range parseErrs {
				tmp = append(tmp, pe.Text)
			}
			text = errTextUnableToParse(errSyntax(strings.Join(tmp, `\n`)))
		}

		return testWarnings(context.Background(), mockup.OptsV2{
			ParseResponse: &v2.ParseResponse{Language: "go", Uast: bufData, Errors: parseErrs},
		}, nil, text, repeats)
	}

	for _, e := range [][]*v2.ParseError{
		testErrParseErrorNilSlice,
		testErrParseErrorOnly,
		testErrParseErrorSeveral,
	} {
		if err := testWarningsParseErrors(e); err != nil {
			return err
		}
	}

	return nil
}

type queryResult struct {
	out string
	err error
}

func testWarnings(ctx context.Context, o mockup.OptsV2, gitbaseEnvs map[string]string, expWarning string, repeats int) error {
	closer, err := mockup.PrepareGRPCServer(mockup.Options{OptsV2: o})
	defer closer()
	if err != nil {
		return err
	}

	test, err := prepareGitBaseEnvironment()
	if err != nil {
		return err
	}

	resc := make(chan queryResult)
	go func() {
		var tmpErr error
		tmpOut := capture_output.Capture(func() {
			tmpErr = test.RunQueryCtx(ctx, gitbaseEnvs, gitbase.Query{
				Statements: []string{uastQuery},
			})
		}, captureOutputDelay)

		resc <- queryResult{out: tmpOut, err: tmpErr}
	}()

	var outPut string
	select {
	case res := <-resc:
		if res.err != nil {
			return res.err
		}
		outPut = res.out
	case <-time.After(queryTimeoutTotal):
		return errors.NewKind("query is being executed longer than expected").New()
	}

	log.Infof("out: %v", outPut)
	actRepeats := strings.Count(outPut, expWarning)
	if repeats != actRepeats {
		return errors.NewKind("repeats of %q: exp: %v act: %v").New(expWarning, repeats, actRepeats)
	}

	return nil
}

func prepareGitBaseEnvironment() (*gitbase.Test, error) {
	config := regression.NewConfig()
	gitServerConfig := regression.GitServerConfig{}

	args, err := parse(&config)
	if err != nil {
		return nil, err
	}
	_, err = parse(&gitServerConfig)
	if err != nil {
		return nil, err
	}

	if config.ShowRepos {
		repos, err := regression.NewRepositories(gitServerConfig)
		if err != nil {
			return nil, errors.NewKind("could not get repositories").Wrap(err)
		}

		repos.ShowRepos()
		os.Exit(0)
	}

	if len(args) == 0 {
		return nil, errors.NewKind("at least one version required").New()
	}
	config.Versions = args

	test, err := gitbase.NewTest(config, gitServerConfig)
	if err != nil {
		return nil, err
	}

	log.Infof("Preparing run")
	if err := test.Prepare(); err != nil {
		return nil, errors.NewKind("could not prepare environment").Wrap(err)
	}

	return test, nil
}

func parse(data interface{}) ([]string, error) {
	parser := flags.NewParser(data, flags.Default)

	args, err := parser.Parse()
	if err != nil {
		if err, ok := err.(*flags.Error); ok {
			if err.Type == flags.ErrHelp {
				os.Exit(0)
			}
		}
		return nil, errors.NewKind("could not parse arguments").Wrap(err)
	}
	return args, nil
}

func getUASTBytes(text string) ([]byte, error) {
	node := nodes.String(text)
	buf := bytes.NewBuffer(nil)
	if err := nodesproto.WriteTo(buf, node); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func errTextUnableToParse(errText string) string {
	return "unable to parse the given blob using bblfsh: " + errText
}

func errTextRPC(errText string) string {
	return "rpc error: code = Unknown desc = " + errText
}

func errSyntax(errText string) string {
	return "syntax error: " + errText
}
