package simulation

import (
	pb "github.com/patrickz98/project.go.omnetpp/proto"
	"github.com/patrickz98/project.go.omnetpp/simple"
	"github.com/patrickz98/project.go.omnetpp/storage"
)

var excludePrefix = []string{
	".git/",
}

func Upload(config *Config) (ref *pb.StorageRef, err error) {

	logger.Println("zipping", config.Path)

	buf, err := simple.TarGz(config.Path, config.SimulationId, excludePrefix...)
	if err != nil {
		return
	}

	logger.Println("uploading", config.SimulationId)

	ref, err = storage.Upload(&buf, storage.FileMeta{
		Bucket:   config.SimulationId,
		Filename: "source.tar.gz",
	})
	if err != nil {
		return
	}

	return
}
