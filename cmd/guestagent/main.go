package main

import (
	"fmt"
	"grove/common"
	"log"
)

func main() {
	output, err := common.ExecuteWithTimeout(1000, true, false, "cmd.exe", "/C", "dir")
	if err != nil {
		log.Fatalf("Error executing command: %v", err)
	}
	fmt.Println(output)
}
