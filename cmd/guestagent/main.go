package main

import (
	"fmt"
	"grove/common"
	"log"
)

func main() {
	args := []string{"ls", "/etc/"}
	kwargs := map[string]interface{}{
		"log_output_on_error": true,
		"run_as_root":         false,
	}
	output, err := common.ExecuteWithTimeout(args, kwargs)
	if err != nil {
		log.Fatalf("Error executing command: %v", err)
	}
	fmt.Println(output)
}
