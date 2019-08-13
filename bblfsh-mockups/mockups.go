package bblfsh_mockups

import (
	"net"
	"os"
	"time"

	cmdutil "github.com/bblfsh/sdk/v3/cmd"
	protocol2 "github.com/bblfsh/sdk/v3/protocol"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"gopkg.in/src-d/go-log.v1"
)

const (
	network = "tcp"
	address = "0.0.0.0:9432"

	defaultMsgSizeMB = 100

	// https://github.com/bblfsh/bblfshd/blob/9b49a8fdabe9b6774d91e02b9a35ccce131816d8/daemon/daemon.go#L29
	keepaliveMinTime           = 1 * time.Minute
	keepalivePingWithoutStream = true
)

// Options represents options of the mocked service
type Options struct {
	OptsV2 OptsV2
}

// PrepareGRPCServer runs GRPC server with a mocked service
func PrepareGRPCServer(o Options) (func(), error) {
	GRPCOpts, err := prepareGRPCOptions()
	if err != nil {
		return func() {}, err
	}
	server := grpc.NewServer(GRPCOpts...)

	s2 := NewServiceV2(o.OptsV2)
	protocol2.RegisterDriverServer(server, s2)
	protocol2.RegisterDriverHostServer(server, s2)

	listener, err := net.Listen(network, address)
	if err != nil {
		return func() { server.Stop() }, err
	}

	go func() {
		if err = server.Serve(listener); err != nil {
			log.Errorf(err, "error starting server")
			os.Exit(1)
		}
	}()

	return func() { server.Stop() }, nil
}

func prepareGRPCOptions() ([]grpc.ServerOption, error) {
	opts := append(protocol2.ServerOptions(),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             keepaliveMinTime,
			PermitWithoutStream: keepalivePingWithoutStream,
		}),
	)
	sizeOpt, err := cmdutil.GRPCSizeOptions(defaultMsgSizeMB)
	if err != nil {
		return nil, err
	}
	return append(opts, sizeOpt...), nil
}
