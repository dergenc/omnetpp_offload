package provider

import (
	"context"
	"github.com/pzierahn/omnetpp_offload/eval"
	"github.com/pzierahn/omnetpp_offload/gconfig"
	pb "github.com/pzierahn/omnetpp_offload/proto"
	"github.com/pzierahn/omnetpp_offload/simple"
	"github.com/pzierahn/omnetpp_offload/stargate"
	"github.com/pzierahn/omnetpp_offload/stargrpc"
	"github.com/pzierahn/omnetpp_offload/storage"
	"github.com/pzierahn/omnetpp_offload/sysinfo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"sync"
	"time"
	"fmt"
)

type simulationId = string

type provider struct {
	pb.UnimplementedProviderServer
	providerId     string
	numJobs        int
	store          *storage.Server
	slots          chan int
	mu             *sync.RWMutex
	sessions       map[simulationId]*pb.Session
	executionTimes map[simulationId]time.Duration
	newRecv        *sync.Cond
	allocRecvs     map[simulationId]chan<- int
}

func TryDial(config gconfig.Config) (brokerConn grpc.ClientConnInterface, err error){
	brokerConn, err = grpc.Dial(
		config.Broker.BrokerDialAddr(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithReturnConnectionError(),
		grpc.WithTimeout(2*time.Second),
	)
	return 
}

func StartServer(ctx context.Context, config gconfig.Config, prov *provider) { 

	//
	// Start stargate-gRPC servers.
	//
	
	server := grpc.NewServer()
	pb.RegisterProviderServer(server, prov)
	pb.RegisterStorageServer(server, prov.store)
	pb.RegisterEvaluationServer(server, &eval.Server{})

	stargate.SetConfig(stargate.Config{
		Addr: config.Broker.Address,
		Port: config.Broker.StargatePort,
	})

	go stargrpc.ServeLocal(prov.providerId, server)
	go stargrpc.ServeP2P(prov.providerId, server)
	go stargrpc.ServeRelay(prov.providerId, server)

	go func() {
		select {
		case <-ctx.Done():
			server.Stop()
		}
	}()

}

func Connect(ctx context.Context, config gconfig.Config, prov *provider) (stream pb.Broker_RegisterClient){

	// Debug web server
	// startWatchers(prov)

	//
	// Register provider
	//

	log.Printf("connect to broker %v", config.Broker.BrokerDialAddr())

	brokerConn, err := TryDial(config)
	if err != nil {
		log.Printf("there might be an error with dialing, will try to reconnect")
		return
	}	

	// TODO: Defer should be discussed later on to avoid unintended persistent connections.
	//defer func() { _ = brokerConn.Close()}()

	broker := pb.NewBrokerClient(brokerConn)
	stream, err = broker.Register(ctx)
	if err != nil {
		log.Fatalln(err)
	}

	_ = stream.Send(&pb.Ping{Cast: &pb.Ping_Register{Register: prov.info()}})
	
	return stream
}

func Start(ctx context.Context, config gconfig.Config) {

	mu := &sync.RWMutex{}
	prov := &provider{
		providerId:     simple.NamedId(config.Provider.Name, 8),
		numJobs:        config.Provider.Jobs,
		store:          &storage.Server{},
		slots:          make(chan int, config.Provider.Jobs),
		mu:             mu,
		newRecv:        sync.NewCond(mu),
		sessions:       make(map[simulationId]*pb.Session),
		executionTimes: make(map[simulationId]time.Duration),
		allocRecvs:     make(map[simulationId]chan<- int),
	}

	log.Printf("start provider (%v)", prov.providerId)

	//
	// Init stuff
	//

	prov.recoverSessions()

	stream := Connect(ctx, config, prov)
	StartServer(ctx, config, prov)

	go func() {

		isDisconnected := false

		for range time.Tick(time.Millisecond * 1000) {
			var util *pb.Utilization
			util, err := sysinfo.GetUtilization(ctx)
			if err != nil {
				log.Fatalln(err)
			}

			// NOTE: This is a workaround since current version of go-grpc cannot detect errors (EOF) on an active stream
			_, err = TryDial(config)
			if err != nil {
				if !isDisconnected {
					//server.GracefulStop()
					isDisconnected = true
				}

				fmt.Println("connection lost, trying to reconnect to broker (%v)", prov.providerId)
				continue
			}
			if isDisconnected {
				stream = Connect(ctx, config, prov)
				isDisconnected = false
			}
			log.Printf("Start: send utilization %v", util.CpuUsage)
			err = stream.Send(&pb.Ping{Cast: &pb.Ping_Util{Util: util}})
			fmt.Println(err)
		}
	}()

	//
	// Start resource allocator.
	//

	prov.startAllocator(config.Provider.Jobs)

	return
}
