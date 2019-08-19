package bblfsh_mockups

import (
	"context"
	"time"

	v2 "github.com/bblfsh/sdk/v3/protocol"
)

// TODO server panics when version is empty

// OptsV2 is a set of ServiceV2 mockup options
type OptsV2 struct {
	ParseResponseLag time.Duration
	ParseResponse    *v2.ParseResponse
	ParseResponseErr error
}

// OptsV1 is a set of ServiceV1 mockup options
type OptsV1 struct{}

// ServiceV2 is bblfsh grpc server v2 mockup
type ServiceV2 struct{ Opts OptsV2 }

// ServiceV1 is bblfsh grpc server v2 mockup
type ServiceV1 struct{ Opts OptsV1 }

// NewServiceV2 is ServiceV2 constructor
func NewServiceV2(opts OptsV2) *ServiceV2 { return &ServiceV2{Opts: opts} }

func (v *ServiceV2) ServerVersion(context.Context, *v2.VersionRequest) (*v2.VersionResponse, error) {
	return &v2.VersionResponse{}, nil
}

func (v *ServiceV2) SupportedLanguages(context.Context, *v2.SupportedLanguagesRequest) (*v2.SupportedLanguagesResponse, error) {
	return &v2.SupportedLanguagesResponse{}, nil
}

func (v *ServiceV2) Parse(ctx context.Context, p *v2.ParseRequest) (*v2.ParseResponse, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(v.Opts.ParseResponseLag):
	}
	return v.Opts.ParseResponse, v.Opts.ParseResponseErr
}
