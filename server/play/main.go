package main

import (
	"context"
	blobpb "coolcar/blob/api/gen/v1"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.Dial(
		"localhost:8084",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		panic(err)
	}

	c := blobpb.NewBlobServiceClient(conn)
	ctx := context.Background()
	// res, err := c.CreateBlob(ctx, &blobpb.CreateBlobRequest{
	// 	AccountId:           "account1",
	// 	UploadUrlTimeoutSec: 1000,
	// })

	res, err := c.GetBlobURL(ctx, &blobpb.GetBlobURLRequest{Id: "624b039afb667cf59d03fb5a", TimeoutSec: 3600})
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v\n", res)
}
