package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Print

var builtins = map[string]bool{
	"exit": true,
	"echo": true,
	"type": true,
}

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

		if command == "" {
			continue
		}

		fields := strings.Fields(command)
		cmd := fields[0]

		if command == "exit" {
			return
		}

		if cmd == "echo" {
			if command == "echo" {
				fmt.Println()
			} else {
				fmt.Println(command[len("echo "):])
			}
			continue
		}

		if cmd == "type" {
			if len(fields) >= 2 {
				if builtins[fields[1]] {
					fmt.Println(fields[1] + " is a shell builtin")
				} else {
					fmt.Println(fields[1] + ": not found")
				}
				continue
			} else {
				continue
			}
		}

		fmt.Println(command + ": command not found")
	}
}
