package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Print

var builtins = map[string]bool{
	"exit": true,
	"echo": true,
	"type": true,
	"pwd":  true,
	"cd":   true,
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
		args := fields[1:]

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
				target := fields[1]
				if builtins[target] {
					fmt.Println(fields[1] + " is a shell builtin")
					continue
				}

				pathenv := os.Getenv("PATH")
				dirs := strings.Split(pathenv, string(os.PathListSeparator))

				found := false

				for _, dir := range dirs {
					fullPath := filepath.Join(dir, target)

					info, err := os.Stat(fullPath)

					if err != nil {
						continue
					}

					if info.Mode().IsRegular() && info.Mode().Perm()&0111 != 0 {
						fmt.Println(target + " is " + fullPath)
						found = true
						break
					}
				}

				if !found {
					fmt.Println(target + ": not found")
				}

				continue
			} else {
				continue
			}
		}

		if cmd == "pwd" {
			wd, err := os.Getwd()
			if err != nil {
				continue
			}

			fmt.Println(wd)
			continue
		}

		if cmd == "cd" {
			if len(args) == 0 {
				continue
			}

			path := args[0]

			if strings.HasPrefix(path, "/") {
				err := os.Chdir(path)
				if err != nil {
					fmt.Println("cd:", path+": No such file or directory")
				}
			}
			continue
		}

		pathenv := os.Getenv("PATH")
		dirs := strings.Split(pathenv, string(os.PathListSeparator))

		executed := false

		for _, dir := range dirs {
			fullPath := filepath.Join(dir, cmd)
			info, err := os.Stat(fullPath)
			if err != nil {
				continue
			}
			if info.Mode().IsRegular() && info.Mode().Perm()&0111 != 0 {
				c := exec.Command(cmd, args...)
				c.Stdin = os.Stdin
				c.Stdout = os.Stdout
				c.Stderr = os.Stderr
				c.Run()
				executed = true
				break
			}
		}

		if !executed {
			fmt.Println(command + ": command not found")
		}
	}
}
