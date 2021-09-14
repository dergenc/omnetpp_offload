package consumer

import (
	"context"
	"github.com/pzierahn/project.go.omnetpp/eval"
	"github.com/pzierahn/project.go.omnetpp/gconfig"
	pb "github.com/pzierahn/project.go.omnetpp/proto"
	"github.com/pzierahn/project.go.omnetpp/simple"
	"google.golang.org/grpc"
	"log"
	"path/filepath"
)

func Start(ctx context.Context, config *Config) {

	//go statisticJsonApi()

	if config.Tag == "" {
		config.Tag = filepath.Base(config.Path)
	}

	id := simple.NamedId(config.Tag, 8)
	log.Println("#################################################")
	log.Printf("Start: simulation %s", id)
	log.Println("#################################################")

	log.Printf("Start: connecting to broker (%v)", gconfig.BrokerDialAddr())

	conn, err := grpc.Dial(
		gconfig.BrokerDialAddr(),
		grpc.WithInsecure(),
		grpc.WithBlock(),
	)
	if err != nil {
		log.Fatalln(err)
	}
	defer func() {
		_ = conn.Close()
	}()

	eval.Init(conn)
	eval.SetScenario(id)
	eval.DeviceId = "consumer"

	//log.Printf("Start: set execution timeout to %v", timeout)

	// TODO: Fix timeout
	//go func() {
	//	// TODO: Find a more elegant way of doing this
	//	<-ctx.Done()
	//	log.Fatalf("Start: execution timeout")
	//}()

	cons := &consumer{
		ctx:    ctx,
		config: config,
		bconn:  conn,
		simulation: &pb.Simulation{
			Id:        id,
			OppConfig: config.OppConfig,
		},
		allocate: newQueue(),
	}

	log.Printf("Start: zipping %s", cons.config.Path)

	done := eval.LogAction(eval.ActionCompress, "source")
	buf, err := simple.TarGz(cons.config.Path, cons.simulation.Id, cons.config.Ignore...)
	_ = done(err)
	if err != nil {
		log.Fatalln(err)
	}

	cons.simulationSource = buf.Bytes()

	onInit := make(chan int32)
	defer close(onInit)

	go cons.startConnector(onInit)

	cons.finished.Add(int(<-onInit))
	cons.finished.Wait()

	log.Printf("Start: simulation finished!")

	//scenarioName := fmt.Sprintf("s%0s-t%0s", eval.ScenarioId, eval.TrailId)
	//eval.WriteRuns(filepath.Join(config.Path, "opp-edge-eval-runs-"+scenarioName+".csv"))
	//eval.WriteTransfers(filepath.Join(config.Path, "opp-edge-eval-transfers-"+scenarioName+".csv"))
	//eval.WriteSetup(filepath.Join(config.Path, "opp-edge-eval-setup-"+scenarioName+".csv"))

	return
}
