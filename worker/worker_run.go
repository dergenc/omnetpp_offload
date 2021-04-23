package worker

import (
	"com.github.patrickz98.omnet/defines"
	"com.github.patrickz98.omnet/omnetpp"
	pb "com.github.patrickz98.omnet/proto"
	"com.github.patrickz98.omnet/simple"
	"com.github.patrickz98.omnet/storage"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

var setupSync sync.Mutex

var copyIgnores = map[string]bool{
	// Don't copy results
	"results/": true,
}

func setup(job *pb.Work) (project omnetpp.OmnetProject, err error) {

	// Prevent that a simulation will be downloaded multiple times
	setupSync.Lock()
	defer setupSync.Unlock()

	// Simulation directory with simulation source code
	simulationBase := filepath.Join(defines.Simulation, job.SimulationId)

	// This will be the working directory, that contains the results for the job
	// A symbolic copy is created to use all configs, ned files and ini files
	simulationPath := filepath.Join(defines.Simulation, "mirrors", simple.NamedId(job.SimulationId, 8))

	if _, err = os.Stat(simulationBase); err == nil {

		//
		// Simulation already downloaded and prepared
		//

		err = simple.SymbolicCopy(simulationBase, simulationPath, copyIgnores)
		if err != nil {
			return
		}

		logger.Printf("simulation %s already downloaded\n", job.SimulationId)
		project = omnetpp.New(simulationPath)

		return
	}

	//
	// Download and compile the simulation
	//

	logger.Printf("checkout %s to %s\n", job.SimulationId, simulationBase)

	byt, err := storage.Download(job.Source)
	if err != nil {
		return
	}

	err = simple.UnTarGz(defines.Simulation, byt)
	if err != nil {
		_ = os.RemoveAll(simulationBase)
		return
	}

	logger.Printf("setup %s\n", job.SimulationId)

	// Compile simulation source code
	srcProject := omnetpp.New(simulationBase)
	err = srcProject.Setup()
	if err != nil {
		return
	}

	// Create a new symbolic copy
	err = simple.SymbolicCopy(simulationBase, simulationPath, copyIgnores)
	if err != nil {
		return
	}

	project = omnetpp.New(simulationPath)

	return
}

func (client *workerConnection) uploadResults(project omnetpp.OmnetProject, job *pb.Work) (err error) {

	buf, err := project.ZipResults()
	if err != nil {
		return
	}

	ref, err := storage.Upload(&buf, storage.FileMeta{
		Bucket:   job.SimulationId,
		Filename: fmt.Sprintf("results_%s_%s.tar.gz", job.Config, job.RunNumber),
	})
	if err != nil {
		return
	}

	results := pb.WorkResult{
		Job:     job,
		Results: ref,
	}

	aff, err := client.client.Push(context.Background(), &results)
	if err != nil {
		// TODO: Delete storage upload
		// _ = storage.Delete(ref)
		return
	}

	if aff.Error != "" {
		err = fmt.Errorf(aff.Error)
	}

	return
}

func (client *workerConnection) runTasks(job *pb.Work) {

	//
	// Setup simulation environment
	// Includes downloading and compiling the simulation
	//

	opp, err := setup(job)
	if err != nil {
		logger.Fatalln(err)
	}

	//
	// Setup simulation environment
	//

	err = opp.RunLog(job.Config, job.RunNumber)
	if err != nil {
		logger.Fatalln(err)
	}

	//
	// Upload simulation results
	//

	err = client.uploadResults(opp, job)
	if err != nil {
		logger.Fatalln(err)
	}

	// Todo: Cleanup simulationBase

	// Cleanup symbolic mirrors
	_ = os.RemoveAll(opp.SourcePath)
}
