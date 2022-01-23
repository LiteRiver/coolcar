package main

import (
	"fmt"
	"time"
)

func main() {
	now := time.Now()
	fmt.Printf("now is: %d\n", now.Unix())
	expiresAt := now.Add(7200 * time.Second)

	fmt.Printf("after 7200s: %d", expiresAt.Unix())
}
