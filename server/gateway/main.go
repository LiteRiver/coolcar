package main

import (
	"context"
	"coolcar/auth/api/gen/v1"
	carpb "coolcar/car/api/gen/v1"
	rentalpb "coolcar/rental/api/gen/v1"
	"coolcar/shared/auth"
	"coolcar/shared/server"
	"github.com/namsral/flag"
	"log"
	"net/http"
	"net/textproto"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
)

var addr = flag.String("addr", ":8081", "address to listen")
var authAddr = flag.String("auth_addr", "localhost:8082", "address for auth service")
var tripAddr = flag.String("trip_addr", "localhost:8083", "address for trip service")
var profileAddr = flag.String("profile_addr", "localhost:8084", "address for profile service")
var carAddr = flag.String("car_addr", "localhost:8085", "address for car service")

func main() {
	flag.Parse()

	logger, err := server.NewZapLogger()
	if err != nil {
		log.Fatalf("cannot create logger: %v", err)
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux(
		runtime.WithMarshalerOption(
			runtime.MIMEWildcard,
			&runtime.JSONPb{
				MarshalOptions: protojson.MarshalOptions{UseEnumNumbers: true, UseProtoNames: true},
			},
		),
		runtime.WithIncomingHeaderMatcher(func(s string) (string, bool) {
			if s == textproto.CanonicalMIMEHeaderKey(runtime.MetadataHeaderPrefix+auth.ImpersonateAccountHeader) {
				return "", false
			}
			return runtime.DefaultHeaderMatcher(s)
		}),
	)

	serverConfigs := []struct {
		name         string
		addr         string
		registerFunc func(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) error
	}{
		{
			name:         "auth",
			addr:         *authAddr,
			registerFunc: authpb.RegisterAuthServiceHandlerFromEndpoint,
		},
		{
			name:         "trip",
			addr:         *tripAddr,
			registerFunc: rentalpb.RegisterTripServiceHandlerFromEndpoint,
		},
		{
			name:         "profile",
			addr:         *profileAddr,
			registerFunc: rentalpb.RegisterProfileServiceHandlerFromEndpoint,
		},
		{
			name:         "car",
			addr:         *carAddr,
			registerFunc: carpb.RegisterCarServiceHandlerFromEndpoint,
		},
	}

	for _, cfg := range serverConfigs {
		err := cfg.registerFunc(ctx, mux, cfg.addr, []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())})

		if err != nil {
			logger.Sugar().Fatalf("cannot register service %s", cfg.name)
		}
	}

	http.HandleFunc("/healthz", func(rw http.ResponseWriter, r *http.Request) {
		rw.Write([]byte("OK"))
	})

	http.Handle("/", mux)

	logger.Sugar().Infof("grpc gateway started at: %s", *addr)
	logger.Sugar().Fatal(http.ListenAndServe(*addr, nil))
}
