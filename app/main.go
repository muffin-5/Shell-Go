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
	"exit":    true,
	"echo":    true,
	"type":    true,
	"pwd":     true,
	"cd":      true,
	"history": true,
}

type commandCompleter struct {
	lastPrefix string
	tabCount   int
}

var historyList []string

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

	builtinList := []string{"echo", "exit", "type", "pwd", "cd", "history"}

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

func runBuiltint(cmd string, args []string) {
	switch cmd {
	case "echo":
		fmt.Println(strings.Join(args, " "))
	case "type":
		if len(args) > 0 {
			if builtins[args[0]] {
				fmt.Println(args[0] + " is a shell builtin")
			} else {
				fmt.Println(args[0] + ": not found")
			}
		}
	case "pwd":
		wd, _ := os.Getwd()
		fmt.Println(wd)
	case "history":
		for i, h := range historyList {
			fmt.Printf("%5d  %s\n", i+1, h)
		}
	}

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

		historyList = append(historyList, command)

		fields := parseCommand(command)

		pipeIndex := -1
		pipeCount := 0
		for i, f := range fields {
			if f == "|" {
				if pipeIndex == -1 {
					pipeIndex = i
				}
				pipeCount++
				// break
			}
		}

		if pipeIndex != -1 {

			if pipeCount > 1 {

				var pipeline [][]string
				var current []string

				for _, f := range fields {
					if f == "|" {
						pipeline = append(pipeline, current)
						current = []string{}
					} else {
						current = append(current, f)
					}

				}
				pipeline = append(pipeline, current)

				n := len(pipeline)
				if n < 2 {
					fmt.Println("invalid pipeline")
					continue
				}

				pipes := make([][2]*os.File, n-1)
				for i := 0; i < n-1; i++ {
					r, w, err := os.Pipe()
					if err != nil {
						fmt.Println("pipe error:", err)
						continue
					}
					pipes[i] = [2]*os.File{r, w}
				}

				var cmds []*exec.Cmd

				for i, args := range pipeline {
					cmd := exec.Command(args[0], args[1:]...)

					if i > 0 {
						cmd.Stdin = pipes[i-1][0]
					} else {
						cmd.Stdin = os.Stdin
					}

					if i < n-1 {
						cmd.Stdout = pipes[i][1]
					} else {
						cmd.Stdout = os.Stdout
					}

					cmd.Stderr = os.Stderr
					cmds = append(cmds, cmd)
				}

				for _, cmd := range cmds {
					if err := cmd.Start(); err != nil {
						fmt.Println(err)
						continue
					}
				}

				for _, p := range pipes {
					p[0].Close()
					p[1].Close()
				}

				for _, cmd := range cmds {
					cmd.Wait()
				}

				continue

			} else {

				cmd1 := fields[:pipeIndex]
				cmd2 := fields[pipeIndex+1:]

				if len(cmd1) == 0 || len(cmd2) == 0 {
					fmt.Println("Invalid pipeline")
				}

				r, w, err := os.Pipe()
				if err != nil {
					fmt.Println("pipe error:", err)
					continue
				}

				if builtins[cmd1[0]] && builtins[cmd2[0]] {
					oldStdout := os.Stdout
					oldStdin := os.Stdin

					os.Stdout = w
					runBuiltint(cmd1[0], cmd1[1:])
					w.Close()

					os.Stdin = r
					runBuiltint(cmd2[0], cmd2[1:])

					os.Stdout = oldStdout
					os.Stdin = oldStdin
					r.Close()
					continue
				}

				if builtins[cmd1[0]] {
					oldStdout := os.Stdout
					os.Stdout = w

					runBuiltint(cmd1[0], cmd1[1:])

					w.Close()
					os.Stdout = oldStdout

					c2 := exec.Command(cmd2[0], cmd2[1:]...)
					c2.Stdin = r
					c2.Stdout = os.Stdout
					c2.Stderr = os.Stderr
					c2.Run()

					r.Close()
					continue
				}

				if builtins[cmd2[0]] {
					c1 := exec.Command(cmd1[0], cmd1[1:]...)
					c1.Stdout = w
					c1.Stderr = os.Stderr

					c1.Start()
					w.Close()

					oldStdin := os.Stdin
					os.Stdin = r

					runBuiltint(cmd2[0], cmd2[1:])

					os.Stdin = oldStdin
					r.Close()

					c1.Wait()
					continue
				}

				c1 := exec.Command(cmd1[0], cmd1[1:]...)
				c2 := exec.Command(cmd2[0], cmd2[1:]...)

				c1.Stdout = w
				c1.Stderr = os.Stderr

				c2.Stdin = r
				c2.Stdout = os.Stdout
				c2.Stderr = os.Stderr

				if err := c1.Start(); err != nil {
					fmt.Println(err)
					continue
				}

				if err := c2.Start(); err != nil {
					fmt.Println(err)
					continue
				}

				w.Close()
				r.Close()

				c1.Wait()
				c2.Wait()
				continue
			}
		}

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

		if cmd == "history" {
			for i, h := range historyList {
				fmt.Printf("%5d %s\n", i+1, h)
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
