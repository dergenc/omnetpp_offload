package broker

import (
	pb "com.github.patrickz98.omnet/proto"
	"com.github.patrickz98.omnet/utils"
	"encoding/json"
	"fmt"
	"google.golang.org/grpc/metadata"
)

func (server *broker) Link(stream pb.Broker_LinkServer) (err error) {

	md, ok := metadata.FromIncomingContext(stream.Context())

	if !ok {
		logger.Printf("metadata missing")
		err = fmt.Errorf("metadata missing")
		return
	}

	var id string
	id, err = utils.MetaString(md, "workerId")
	if err != nil {
		return
	}

	logger.Println("linked", id)

	//
	// Send work to clients
	//

	work := make(chan *pb.Tasks)
	defer func() {
		server.queue.Unlink(id)
		close(work)
	}()

	server.queue.Link(id, work)

	go func() {
		for {
			job, ok := <-work
			if !ok {
				logger.Println("exit work mode for", id)
				break
			}

			logger.Println("send work to", id)

			err := stream.Send(job)
			if err != nil {
				logger.Println(err)
				break
			}
		}
	}()

	//
	// Receive Client Info
	//

	for {
		var info *pb.ResourceCapacity
		info, err = stream.Recv()
		if err != nil {
			break
		}

		if id == "" {
			id = info.WorkerId
		}

		server.workers.Put(info.WorkerId, info)

		jsonBytes, _ := json.MarshalIndent(info, "", "    ")
		logger.Println("link", string(jsonBytes))

		server.distributeWork()
	}

	logger.Println("unlinked", id)

	return
}
