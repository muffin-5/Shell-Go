package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/chzyer/readline"
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

type commandCompleter struct {
	lastPrefix string
	tabCount   int
}

func longestCommonPrefix(strs []string) string {
	if len(strs) == 0 {
		return ""
	}

	prefix := strs[0]
	for _, s := range strs[1:] {
		for !strings.HasPrefix(s, prefix) {
			if len(prefix) == 0 {
				return ""
			}
			prefix = prefix[0 : len(prefix)-1]
		}
	}

	return prefix
}
func (c *commandCompleter) Do(line []rune, pos int) (newLine [][]rune, length int) {

	lineStr := string(line[:pos])
	trimmedStr := strings.TrimLeft(lineStr, " ")

	if strings.Contains(trimmedStr, " ") {
		return nil, 0
	}

	if trimmedStr == c.lastPrefix {
		c.tabCount++
	} else {
		c.lastPrefix = trimmedStr
		c.tabCount = 1
	}

	seen := make(map[string]bool)
	var matches []string

	addCandidate := func(name string) {
		if seen[name] {
			return
		}

		if strings.HasPrefix(name, trimmedStr) {
			seen[name] = true
			matches = append(matches, name)
		}
	}

	builtinList := []string{"echo", "exit", "type", "pwd", "cd"}

	for _, cmd := range builtinList {
		addCandidate(cmd)
	}

	pathEnv := os.Getenv("PATH")
	dirs := strings.Split(pathEnv, string(os.PathListSeparator))

	for _, dir := range dirs {
		files, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, file := range files {
			if strings.HasPrefix(file.Name(), trimmedStr) {
				info, err := file.Info()
				if err != nil {
					continue
				}

				if info.Mode().IsRegular() && info.Mode().Perm()&0111 != 0 {
					addCandidate(file.Name())
				}
			}
		}
	}

	if len(matches) == 0 {
		fmt.Print("\x07")
		return nil, 0
	}

	if len(matches) == 1 {
		suffix := matches[0][len(trimmedStr):] + " "
		c.tabCount = 0
		return [][]rune{[]rune(suffix)}, len(trimmedStr)
	}

	lcp := longestCommonPrefix(matches)
	if len(lcp) > len(trimmedStr) {
		suffix := lcp[len(trimmedStr):]
		c.tabCount = 0
		return [][]rune{[]rune(suffix)}, len(trimmedStr)
	}

	if c.tabCount == 1 {
		fmt.Print("\x07")
		return nil, 0
	}

	sort.Strings(matches)

	fmt.Print("\n")
	fmt.Print(strings.Join(matches, "  "))
	fmt.Print("\n")
	fmt.Print("$ " + trimmedStr)

	c.tabCount = 0
	return nil, 0
}

func extractRedirection(args []string) ([]string, string, bool, string, bool) {
	var stdoutFile string
	var stderrFile string
	appendStdout := false
	appendStderr := false

	newArgs := []string{}

	for i := 0; i < len(args); i++ {
		if i+1 < len(args) {
			switch args[i] {
			case ">", "1>":
				stdoutFile = args[i+1]
				appendStdout = false
				i++
				continue

			case ">>", "1>>":
				stdoutFile = args[i+1]
				appendStdout = true
				i++
				continue
			case "2>":
				stderrFile = args[i+1]
				appendStderr = false
				i++
				continue
			case "2>>":
				stderrFile = args[i+1]
				appendStderr = true
				i++
				continue
			}
		}
		newArgs = append(newArgs, args[i])

	}
	return newArgs, stdoutFile, appendStdout, stderrFile, appendStderr
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

	rl, err := readline.NewEx(&readline.Config{
		Prompt:       "$ ",
		AutoComplete: &commandCompleter{},
		Stdin:        os.Stdin,
		Stdout:       os.Stdout,
	})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer rl.Close()

	for {

		command, err := rl.Readline()
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

		if cmd == "exit" {
			return
		}

		if cmd == "echo" {
			args, outFile, appendOut, errFile, appendErr := extractRedirection(args)
			output := strings.Join(args, " ")

			if errFile != "" {
				flags := os.O_CREATE | os.O_WRONLY
				if appendErr {
					flags |= os.O_APPEND
				} else {
					flags |= os.O_TRUNC
				}
				f, err := os.OpenFile(errFile, flags, 0644)
				if err != nil {
					f.Close()
				}
			}

			if outFile != "" {
				flags := os.O_CREATE | os.O_WRONLY
				if appendOut {
					flags |= os.O_APPEND
				} else {
					flags |= os.O_TRUNC
				}
				f, err := os.OpenFile(outFile, flags, 0644)
				if err != nil {
					fmt.Println("error opening file:", err)
					continue
				}
				fmt.Fprintln(f, output)
				f.Close()

			} else {
				fmt.Println(output)
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

		args, outFile, appendOut, errFile, appendErr := extractRedirection(args)

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

				if outFile != "" {
					flags := os.O_CREATE | os.O_WRONLY

					if appendOut {
						flags |= os.O_APPEND
					} else {
						flags |= os.O_TRUNC
					}

					f, err := os.OpenFile(outFile, flags, 0644)
					if err != nil {
						fmt.Println("error opening file:", err)
						break
					}

					c.Stdout = f
					defer f.Close()
				} else {
					c.Stdout = os.Stdout
				}

				if errFile != "" {
					flags := os.O_CREATE | os.O_WRONLY

					if appendErr {
						flags |= os.O_APPEND
					} else {
						flags |= os.O_TRUNC
					}

					ef, err := os.OpenFile(errFile, flags, 0644)
					if err != nil {
						fmt.Println("error opening file:", err)
						break
					}

					c.Stderr = ef
					defer ef.Close()
				} else {
					c.Stderr = os.Stderr
				}

				c.Run()
				executed = true
				break
			}
		}

		if !executed {
			fmt.Println(cmd + ": command not found")
		}
	}
}
