package broker

import (
	"context"
	pb "github.com/patrickz98/project.go.omnetpp/proto"
	"google.golang.org/protobuf/proto"
)

func (server *broker) Create(_ context.Context, simulation *pb.Simulation) (resp *pb.SimulationId, err error) {

	sState := server.simulations.createNew(simulation)
	logger.Printf("created new simulation: id='%s'", sState.simulationId)

	resp = &pb.SimulationId{
		Id: sState.simulationId,
	}

	return
}

func (server *broker) GetSimulation(_ context.Context, req *pb.SimulationId) (simulation *pb.Simulation, err error) {

	logger.Printf("GetSimulation: id='%s'", req.Id)
	sState := server.simulations.getSimulationState(req.Id)

	simulation = &pb.Simulation{
		SimulationId: sState.simulationId,
		OppConfig:    sState.oppConfig,
	}

	return
}

func (server *broker) AddTasks(_ context.Context, tasks *pb.Tasks) (resp *pb.Empty, err error) {

	logger.Printf("simulation %s (added %d tasks)", tasks.SimulationId, len(tasks.Items))

	sState := server.simulations.getSimulationState(tasks.SimulationId)
	sState.write(func() {
		for _, task := range tasks.Items {
			id := tId(task)
			sState.queue[id] = true
			sState.runs[id] = task
		}
	})

	resp = &pb.Empty{}

	return
}

func (server *broker) SetSource(_ context.Context, ref *pb.Source) (resp *pb.Empty, err error) {

	logger.Printf("set source for %s to %v", ref.SimulationId, ref.Source)

	sState := server.simulations.getSimulationState(ref.SimulationId)
	sState.write(func() {
		sState.source = ref.Source
	})

	resp = &pb.Empty{}

	return
}

func (server *broker) GetSource(_ context.Context, sim *pb.SimulationId) (resp *pb.Source, err error) {

	logger.Printf("get source for %s", sim.Id)

	sState := server.simulations.getSimulationState(sim.Id)
	sState.read(func() {
		resp = &pb.Source{
			SimulationId: sState.simulationId,
			Source:       sState.source,
		}
	})

	return
}

func (server *broker) AddBinary(_ context.Context, binary *pb.Binary) (resp *pb.Empty, err error) {

	logger.Printf("new binary (%s_%s)", binary.SimulationId, binary.Arch)

	sState := server.simulations.getSimulationState(binary.SimulationId)
	sState.write(func() {
		sState.binaries[osArchId(binary.Arch)] = binary
	})

	server.providers.RLock()
	for _, prov := range server.providers.provider {
		prov.RLock()

		if (prov.building == binary.SimulationId) && (osArchId(binary.Arch) == osArchId(prov.arch)) {
			logger.Printf("%s: remove building ref from %s", binary.SimulationId, prov.id)
			prov.building = ""
		}
		prov.RUnlock()
	}
	server.providers.RUnlock()

	resp = &pb.Empty{}

	return
}

func (server *broker) GetBinary(_ context.Context, binary *pb.SimulationBinaryRequest) (resp *pb.Binary, err error) {

	logger.Printf("get binary (%s_%s)", binary, binary.Arch)

	sState := server.simulations.getSimulationState(binary.SimulationId)
	sState.read(func() {
		resp = sState.binaries[osArchId(binary.Arch)]
	})

	return
}

func (server *broker) GetOppConfig(_ context.Context, simulation *pb.SimulationId) (config *pb.OppConfig, err error) {

	sState := server.simulations.getSimulationState(simulation.Id)
	sState.read(func() {
		config = proto.Clone(sState.oppConfig).(*pb.OppConfig)
	})

	return
}
