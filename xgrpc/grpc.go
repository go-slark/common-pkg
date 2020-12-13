package xgrpc

import (
	"context"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/keepalive"
	"net"
	"time"
)

// RPCServer is RPC server config.
type RPCServer struct {
	Network           string
	Addr              string
	Timeout           time.Duration
	IdleTimeout       time.Duration
	MaxLifeTime       time.Duration
	ForceCloseWait    time.Duration
	KeepAliveInterval time.Duration
	KeepAliveTimeout  time.Duration
}

//f is register server function,在业务中实现注册; server实现
//close graceful: srv.GracefulStop()
func NewGrpcServer(c RPCServer, server interface{}, f func(s *grpc.Server, server interface{})) (*grpc.Server, error) {
	keepParams := grpc.KeepaliveParams(keepalive.ServerParameters{
		MaxConnectionIdle:     time.Duration(c.IdleTimeout),
		MaxConnectionAgeGrace: time.Duration(c.ForceCloseWait),
		Time:                  time.Duration(c.KeepAliveInterval),
		Timeout:               time.Duration(c.KeepAliveTimeout),
		MaxConnectionAge:      time.Duration(c.MaxLifeTime),
	})
	srv := grpc.NewServer(keepParams)
	//pb.RegisterLogicServer(srv, &server{l})
	f(srv, server)
	lis, err := net.Listen(c.Network, c.Addr)
	if err != nil {
		logrus.Errorf("start grpc listen fail, addr:%v, err:%+v", c.Addr, err)
		return srv, errors.WithStack(err)
	}

	go func() {
		if err := srv.Serve(lis); err != nil {
			logrus.Errorf("start grpc server fail, err:%+v", err)
		}
	}()
	return srv, nil
}

// RPCClient is RPC client config.
type RPCClient struct {
	Dial    time.Duration
	Timeout time.Duration
}

const (
	minServerHeartbeat = time.Minute * 10
	maxServerHeartbeat = time.Minute * 30
	// grpc options
	grpcInitialWindowSize     = 1 << 24
	grpcInitialConnWindowSize = 1 << 24
	grpcMaxSendMsgSize        = 1 << 24
	grpcMaxCallMsgSize        = 1 << 24
	grpcKeepAliveTime         = time.Second * 10
	grpcKeepAliveTimeout      = time.Second * 3
	grpcBackoffMaxDelay       = time.Second * 3
)

//client conn实现pb interface, client call
//invoke
func NewGrpcClient(c RPCClient, dns string) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.Dial))
	defer cancel()
	//target is grpc dns
	conn, err := grpc.DialContext(ctx, dns,
		[]grpc.DialOption{
			grpc.WithInsecure(),
			grpc.WithInitialWindowSize(grpcInitialWindowSize),
			grpc.WithInitialConnWindowSize(grpcInitialConnWindowSize),
			grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(grpcMaxCallMsgSize)),
			grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(grpcMaxSendMsgSize)),
			grpc.WithBackoffMaxDelay(grpcBackoffMaxDelay),
			grpc.WithKeepaliveParams(keepalive.ClientParameters{
				Time:                grpcKeepAliveTime,
				Timeout:             grpcKeepAliveTimeout,
				PermitWithoutStream: true,
			}),
			grpc.WithBalancerName(roundrobin.Name),
		}...)
	if err != nil {
		logrus.Errorf("grpc client dial fail, err:%+v", err)
		return conn, errors.WithStack(err)
	}
	return conn, nil
}
