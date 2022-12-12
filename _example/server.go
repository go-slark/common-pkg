package _example

import (
	"context"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	"github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/sirupsen/logrus"
	"github.com/smallfish-root/common-pkg/xgin"
	"github.com/smallfish-root/common-pkg/xgrpc"
	"github.com/smallfish-root/common-pkg/xtransport"
	"github.com/smallfish-root/common-pkg/xtransport/grpc"
	"github.com/smallfish-root/common-pkg/xtransport/http"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
	std_grpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var grpcClient *grpc.GRPCClient

func NewGRPCClient() *grpc.GRPCClient {
	objs := []*grpc.ClientObj{
		&grpc.ClientObj{
			Name: "",
			Addr: "",
		},
	}

	f := func() []std_grpc.DialOption {
		retryOps := []grpc_retry.CallOption{
			grpc_retry.WithMax(3),
			grpc_retry.WithPerRetryTimeout(time.Second * 2),
			grpc_retry.WithBackoff(grpc_retry.BackoffLinearWithJitter(time.Second/2, 0.2)),
		}
		retry := grpc_retry.UnaryClientInterceptor(retryOps...)
		opts := []std_grpc.DialOption{
			std_grpc.WithChainUnaryInterceptor(retry, xgrpc.UnaryClientTimeout(3*time.Second)),
			std_grpc.WithTransportCredentials(insecure.NewCredentials()),
			std_grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
			std_grpc.WithKeepaliveParams(keepalive.ClientParameters{
				Time:                10 * time.Second,
				Timeout:             time.Second,
				PermitWithoutStream: true,
			}),
		}
		return opts
	}

	grpcClient = grpc.NewGRPCClient(objs, f, grpc.WithTimeout(5*time.Second))
	return grpcClient
}

func ChainUnaryInterceptor() []std_grpc.UnaryServerInterceptor {
	entry := logrus.NewEntry(logrus.StandardLogger())
	opt := []grpc_logrus.Option{
		grpc_logrus.WithLevels(func(code codes.Code) logrus.Level {
			if code == codes.OK {
				return logrus.InfoLevel
			}
			return logrus.ErrorLevel
		}),
		grpc_logrus.WithMessageProducer(grpc_logrus.DefaultMessageProducer),
	}

	unary := []std_grpc.UnaryServerInterceptor{
		grpc_middleware.ChainUnaryServer(
			grpc_recovery.UnaryServerInterceptor(),
			grpc_logrus.UnaryServerInterceptor(entry, opt...),
			grpc_logrus.PayloadUnaryServerInterceptor(entry, func(ctx context.Context, fullMethodName string, servingObject interface{}) bool { return true }),
		),
		xgrpc.UnaryServerTimeout(3 * time.Second),
		xgrpc.UnaryServerTracIDInterceptor(),
	}
	return unary
}

func ServerOption() []std_grpc.ServerOption {
	option := []std_grpc.ServerOption{
		std_grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle:     15 * time.Second,
			MaxConnectionAge:      30 * time.Second,
			MaxConnectionAgeGrace: 5 * time.Second,
			Time:                  5 * time.Second,
			Timeout:               1 * time.Second,
		}),
		std_grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             5 * time.Second,
			PermitWithoutStream: true,
		}),
	}
	return option
}

func InitEngine(routers []func(group gin.IRouter)) *gin.Engine {
	c := cors.New(cors.Config{
		AllowOriginFunc:  func(origin string) bool { return true },
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Accept-Encoding", "Authorization", "X-CSRF-Token", "X-Authorization", "Content-Disposition"},
		ExposeHeaders:    []string{"X-Authorization", "Content-Disposition"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
	return xgin.SetEngine(xgin.EngineParam{
		Env:          "",
		BaseUrl:      "/api/v1",
		Routers:      routers,
		HandlerFunc:  []gin.HandlerFunc{c},
		AccessLog:    viper.GetBool("log.access"),
		ExcludePaths: viper.GetStringSlice("log.paths"),
	})
}

func InitRouter() []func(group gin.IRouter) {
	routers := []func(group gin.IRouter){
		func(g gin.IRouter) {
			// TODO
		},
	}
	return routers
}

func RunServer() {
	gClient := NewGRPCClient()
	defer func() {
		gClient.Stop()
	}()

	_ = NewKafka().Consume()

	engine := InitEngine(InitRouter())
	srv := &grpc.RegisterObj{}
	servers := []transport.Server{
		http.NewServer(http.Address("server.addr"), http.Handler(engine)),
		srv.NewGRPCServer(
			grpc.Address("grpc.addr"),
			grpc.ServerOptions(ServerOption()),
			grpc.UnaryInterceptor(ChainUnaryInterceptor()),
		),
	}
	c := make(chan os.Signal, 1)
	eg, ctx := errgroup.WithContext(context.TODO())
	for _, server := range servers {
		s := server
		eg.Go(func() error {
			return s.Start()
		})

		eg.Go(func() error {
			<-c
			cx, cancel := context.WithTimeout(ctx, 3*time.Second)
			defer cancel()
			return s.Stop(cx)
		})
	}

	signal.Notify(c, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT, syscall.SIGSEGV)
	<-c
	close(c)

	err := eg.Wait()
	if err != nil {
		fmt.Printf("eg wait all goroutine exit, err:%+v\n", err)
	}
}
