package main

import (
	"context"
	"fmt"
	"time"
)

type paramKey struct{}

func main() {
	ctx := context.WithValue(context.Background(), paramKey{}, "abc")
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	mainTask(ctx)
}

func mainTask(ctx context.Context) {
	fmt.Printf("main task started with param %q\n", ctx.Value(paramKey{}))
	smallTask(context.Background(), "task1")
	smallTask(ctx, "task2")
}

func smallTask(ctx context.Context, name string) {
	fmt.Printf("%s started with param %q\n", name, ctx.Value(paramKey{}))
	select {
	case <- time.After(6 * time.Second):
		fmt.Printf("%s done\n", name)
	case <- ctx.Done():
		fmt.Printf("%s canceled\n", name)

	}
	ctx.Done()
}
