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

func parseCommand(input string) []string {
	var args []string
	var current strings.Builder
	inSingleQuotes := false
	inDoubleQuotes := false

	for i := 0; i < len(input); i++ {
		c := input[i]

		if c == '\'' && !inDoubleQuotes {
			inSingleQuotes = !inSingleQuotes
			continue
		}

		if c == '"' && !inSingleQuotes {
			inDoubleQuotes = !inDoubleQuotes
			continue
		}

		if c == ' ' && !inSingleQuotes && !inDoubleQuotes {
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
			continue
		}

		if c == '\\' {
			if !inSingleQuotes && !inDoubleQuotes {
				if i < len(input)-1 {
					i++
					c = input[i]
				} else {
					continue
				}
			}

			if inDoubleQuotes {
				if i+1 < len(input) {
					if input[i+1] == '\\' || input[i+1] == '"' {
						current.WriteByte(input[i+1])
						i++
						continue
					}
				}
			}
		}

		current.WriteByte(c)
	}

	if current.Len() > 0 {
		args = append(args, current.String())
	}
	return args
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

		fields := parseCommand(command)

		cmd := fields[0]
		args := fields[1:]

		if command == "exit" {
			return
		}

		if cmd == "echo" {
			fmt.Println(strings.Join(args, " "))
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

			if path == "~" {
				home := os.Getenv("HOME")
				if home == "" {
					fmt.Println("cd: ~: No such file or directory")
				}
				path = home
			}

			err := os.Chdir(path)
			if err != nil {
				fmt.Println("cd:", path+": No such file or directory")
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
