package main

import (
	"fmt"
	"os"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	OSS_ID := os.Getenv("OSS_ID")
	OSS_SECRETS := os.Getenv("OSS_SECRETS")

	client, err := oss.New("https://oss-cn-beijing.aliyuncs.com", OSS_ID, OSS_SECRETS)
	if err != nil {
		panic(err)
	}

	lsRes, err := client.ListBuckets()
	if err != nil {
		panic(err)
	}

	for _, bucket := range lsRes.Buckets {
		fmt.Println("Buckets:", bucket.Name)
	}
}
