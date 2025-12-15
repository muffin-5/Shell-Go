package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Print

func main() {

	reader := bufio.NewReader(os.Stdin)

	for {
		// TODO: Uncomment the code below to pass the first stage
		fmt.Print("$ ")

		//Print invlid command
		command, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Err:", err)
			return
		}

		command = strings.TrimSpace(command)

		if command == "exit" {
			return
		}

		if command == "echo" || strings.HasPrefix(command, "echo ") {
			if command == "echo" {
				fmt.Println()
			} else {
				fmt.Println(command[len("echo "):])
			}
			continue
		}

		fmt.Println(command + ": command not found")
	}
}
