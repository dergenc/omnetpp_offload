package storage

import (
	"bytes"
	pb "com.github.patrickz98.omnet/proto"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"io"
	"time"
)

func Download(file *pb.StorageRef) (byt io.Reader, err error) {

	conn, err := grpc.Dial(storageAddress, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		err = fmt.Errorf("did not connect: %v", err)
		return
	}
	defer func() { _ = conn.Close() }()

	client := pb.NewStorageClient(conn)
	stream, err := client.Get(context.Background(), file)
	if err != nil {
		return
	}

	var buf bytes.Buffer

	start := time.Now()

	for {
		parcel, err := stream.Recv()

		if err == io.EOF {
			break
		}

		if err != nil {
			logger.Fatalln(err)
		}

		_, err = buf.Write(parcel.Payload)
		if err != nil {
			logger.Fatalln(err)
		}
	}

	logger.Printf("received data in %v", time.Now().Sub(start))

	byt = &buf

	return
}
