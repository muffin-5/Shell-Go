# POSIX-Compliant Shell in Go

A custom POSIX-compliant command-line shell built from scratch in Go.  
This shell supports executing external programs, built-in commands, and provides
an interactive REPL similar to Unix shells.

This project demonstrates low-level systems programming concepts such as process
creation, command parsing, and standard I/O handling.

---

## 🚀 Features

- Interactive REPL (Read–Eval–Print Loop)
- Execution of external commands
- Built-in commands:
  - `cd`
  - `pwd`
  - `echo`
- Command parsing and argument handling
- Environment variable support
- Proper exit status handling
- Command piping (|)
- Input/output redirection (>, <)

---

## 🛠 Tech Stack

- **Language:** Go
- **Concepts Used:**
  - Process management (`os/exec`)
  - File descriptors & standard I/O
  - String parsing
  - Error handling
  - REPL design

---

## ▶️ How to Run Locally

### Prerequisites
- Go 1.25 or higher

### Steps
```bash
git clone https://github.com/muffin-5/Shell-Go.git
cd Shell-Go
./your_program.sh
