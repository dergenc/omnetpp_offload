package storage

import (
	"context"
	pb "github.com/patrickz98/project.go.omnetpp/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"io"
	"time"
)

func Upload(data io.Reader, meta FileMeta) (ref *pb.StorageRef, err error) {

	conn, err := grpc.Dial(storageAddress, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return
	}

	defer func() { _ = conn.Close() }()

	md := metadata.New(map[string]string{
		"bucket":   meta.Bucket,
		"filename": meta.Filename,
	})

	ctx := context.Background()
	ctx = metadata.NewOutgoingContext(ctx, md)

	client := pb.NewStorageClient(conn)
	stream, err := client.Put(ctx)
	if err != nil {
		return
	}

	start := time.Now()

	for chunk := range streamReader(data) {
		parcel := pb.StorageParcel{
			Size:    int32(chunk.size),
			Offset:  int64(chunk.offset),
			Payload: chunk.payload,
		}

		err = stream.Send(&parcel)
		if err != nil {
			logger.Fatalln(err)
		}
	}

	ref, err = stream.CloseAndRecv()
	if err != nil {
		return
	}

	logger.Printf("upload %v->%v in %v\n", meta.Bucket, meta.Filename, time.Now().Sub(start))

	return
}
