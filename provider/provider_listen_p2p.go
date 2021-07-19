package provider

import (
	"context"
	pb "github.com/pzierahn/project.go.omnetpp/proto"
	"github.com/pzierahn/project.go.omnetpp/stargate"
	"google.golang.org/grpc"
	"log"
	"net"
)

func (prov *provider) listenP2P() {
	for {
		log.Println("listenP2P: waiting for connect")

		ctx := context.Background()
		p2p, err := stargate.DialQUICgRPCListener(ctx, prov.providerId)
		if err != nil {
			log.Println(err)
			continue
		}

		log.Printf("listenP2P: new connection addr=%v", p2p.Addr())

		go func(p2p net.Listener) {

			// TODO: Find a way to close the p2p connection properly
			defer func() { _ = p2p.Close() }()

			server := grpc.NewServer()
			pb.RegisterProviderServer(server, prov)
			pb.RegisterStorageServer(server, prov.store)
			err := server.Serve(p2p)
			if err != nil {
				log.Println(err)
			}
		}(p2p)
	}
}
