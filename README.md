# ptywrapper

`ptywrapper` is a Go module that provides a simplified interface for running commands in a pseudo-terminal (PTY). It is essentially a wrapper around the [github.com/creack/pty](https://github.com/creack/pty) module, designed to make it easier to implement in your programs.

This module was created to address specific challenges I encountered during development of programs written in Go that needed to use pseudo-terminals, such as blocking operations, unsynchronized copying operations, and issues with user input. `ptywrapper` offers a solution (at least, partial) to these problems, making it easier to run terminal commands within Go programs. It provides features like context support, output cleanup, and easy handling of command exit codes, all while ensuring smooth operation.

## Features

`ptywrapper` provides a number of features to simplify the process of running commands in a pseudo-terminal (PTY):

- **Command Execution in PTY**: `ptywrapper` allows you to easily run any command in a PTY. This can be useful for running commands that require a terminal environment.

- **Context Support**: Each command run in a PTY has an associated context. This allows for better control over the command execution and can be used to cancel the command if necessary.

- **Output Cleanup**: `ptywrapper` automatically cleans up the command's output by removing ANSI escape sequences and certain special characters. This makes the output easier to use in your Go programs.

- **Output Discarding**: You can choose to discard the command's output, which means it won't be printed to the standard output during execution. However, the output will still be captured and stored in a variable for later use.

- **Exit Code Handling**: `ptywrapper` captures the exit code of the command, which can be used to determine whether the command completed successfully.

- **Custom Environment**: `ptywrapper` allows you to specify a custom environment for the command being run. This can be useful if you need to set specific environment variables or modify the existing environment in some way. If no custom environment is provided, the command will be run with the current environment (as returned by `os.Environ()`).

- **Testing Support**: `ptywrapper` includes test files in the `./tests` subfolder that demonstrate how to use the module and can be used for testing purposes.

## Installation

To install `ptywrapper`, you can use `go get`:

```bash
go get github.com/fearlessdots/ptywrapper
```

## Usage

Here's a basic example of how to use ptywrapper:

```go
import (
  "github.com/fearlessdots/ptywrapper"
  "fmt"
)

func main() {
  cmd := &ptywrapper.Command{
    Entry: "ls",
    Args: []string{"-l"},
  }

  completedCmd, err := cmd.RunInPTY()
  if err != nil {
    fmt.Println("Error:", err)
    return
  }

  if completedCmd.ExitCode != 0 {
    fmt.Println("Command exited with error. Exit code:", completedCmd.ExitCode)
  } else {
    fmt.Println("Command output:", completedCmd.Output)
  }
}
```

If you want to run a command but discard its output (i.e., not print it to the standard output), you can set the `Discard` field to true:

```go
package main

import (
  "github.com/fearlessdots/ptywrapper"
  "fmt"
)

func main() {
  cmd := &ptywrapper.Command{
    Entry: "ls",
    Args: []string{"-l"},
    Discard: true, // Discard output
  }

  completedCmd, err := cmd.RunInPTY()
  if err != nil {
    fmt.Println("Error:", err)
    return
  }

  // The output will not be printed to the standard output,
  // but it will still be available in the Output field.
  fmt.Println("Command output:", completedCmd.Output)
}
```

However, it's important to note that even when the output is discarded in this way, it is still captured and stored in the `Output` field of the `Command` struct. This allows you to access and operate on the command's output later in your code, even if it was not immediately visible during execution.

This feature can be particularly useful when you want to run a command silently (without printing its output), but still need to use the output for further processing or logging.

To append custom environment variables to the current environment and use them with a command, the following code can be used:

```go
import (
  "github.com/fearlessdots/ptywrapper"
  "fmt"
  "os"
)

func main() {
  // Get the current environment
  currentEnv := os.Environ()

  // Define custom environment variables
  customEnv := map[string]string{
    "FOO": "bar",
    "BAZ": "qux",
  }

  // Append custom environment variables to the current environment
  for key, value := range customEnv {
    currentEnv = append(currentEnv, key+"="+value)
  }

  cmd := &ptywrapper.Command{
    Entry: "printenv",
    Args: []string{},
    Env: currentEnv,
  }

  completedCmd, err := cmd.RunInPTY()
  if err != nil {
    fmt.Println("Error:", err)
    return
  }

  fmt.Println("Command output:", completedCmd.Output)
}
```

Finally, here's an example of how to use only the custom environment:

```go
import (
  "github.com/fearlessdots/ptywrapper"
  "fmt"
)

func main() {
  cmd := &ptywrapper.Command{
    Entry: "printenv",
    Args: []string{},
    Env: []string{"FOO=bar", "BAZ=qux"},
  }

  completedCmd, err := cmd.RunInPTY()
  if err != nil {
    fmt.Println("Error:", err)
    return
  }

  fmt.Println("Command output:", completedCmd.Output)
}
```

### The `Command` Struct

The `Command` struct represents a command to be run in a pseudo-terminal (PTY). It has several fields that can be set before running the command and some that are populated after the command has been run.

#### Fields Available Before Running `command.RunInPTY()`

- `Entry`: This is the command to be run. It should be a string representing the path to the executable.

- `Args`: This is an array of strings representing the arguments to be passed to the command.

- `Env`: This is an array of strings representing the environment variables for the command. Each string should be in the format `KEY=value`. If `Env` is not set, the command will be run with the current environment (as returned by `os.Environ()`).

- `Discard`: This is a boolean that determines whether the command's output should be discarded. If set to `true`, the command's output will not be printed to the standard output during execution, but it will still be captured and stored in the `Output` field.

#### Fields Available After Running `command.RunInPTY()`

- `Completed`: This is a boolean that indicates whether the command has completed. It is set to `true` after the command has been run.

- `Output`: This is a string that contains the command's output. It is populated after the command has been run. If the `Discard` field was set to `true`, the output will not have been printed to the standard output, but it will still be captured and stored in this field.

- `ExitCode`: This is an integer that contains the command's exit code. It is populated after the command has been run. An exit code of 0 usually indicates that the command completed successfully, while a non-zero exit code usually indicates that an error occurred.

#### Persistence of Fields in the `Command` Struct

After executing `command.RunInPTY()`, the fields that were available before running the command (`Entry`, `Args`, `Env`, and `Discard`) remain accessible. They retain the values that were set before the command was run.

This means you can still access the original command (`Entry`), its arguments (`Args`), the environment variables (`Env`), and the discard setting (`Discard`) even after the command has been executed. These fields are not modified by the execution of the command.

In addition to these, the fields that are populated after the command has been run (`Completed`, `Output`, and `ExitCode`) are also available. This allows you to access a comprehensive set of information about the command and its execution, including what the command was, what arguments it was run with, what environment variables it used, whether its output was discarded, whether it has completed, what output it produced, and what its exit code was.

## Inner Workings

The `ptywrapper` module is composed of several key components:

### Types

- `contextWrapper`: This type is a struct that wraps a context and its cancel function. It is used to track the execution of the command.

- `Writer`: This type is a struct that wraps two file pointers (source and destination) and a context. It implements the `io.Writer` interface and is used to copy data between the source and destination.

- `Command`: This type is a struct that represents a command to be run in a PTY. It includes fields for the command entry, arguments, environment variables, a flag to discard output, a flag to indicate if the command has completed, the command output, and the exit code.

### Functions

- `cleanupString`: This function takes a string as input and cleans it up by removing ANSI escape sequences and certain special characters.

- `generateContextWrapper`: This function generates a new `contextWrapper`.

- `RunInPTY`: This method of the `Command` type runs the command in a PTY. It handles the creation of the PTY, setting up the command, starting the command, copying data between the PTY and the standard input/output, waiting for the command to exit, and cleaning up the command output.

### Goroutines

The `RunInPTY` method uses several goroutines to handle different aspects of the command execution:

- A goroutine is used to resize the PTY whenever a `SIGWINCH` signal is received.

- A goroutine is used to copy data from the standard input to the PTY. This goroutine reads data from the standard input in a non-blocking manner and writes it to the PTY.

- A goroutine is used to copy data from the PTY to the standard output. This goroutine reads data from the PTY in a non-blocking manner and writes it to the standard output and a bytes buffer.

- A goroutine is used to wait for the command to exit. When the command exits, this goroutine cancels the context, closes the PTY, and waits for the other goroutines to finish.

These goroutines work together to ensure that the command is executed in a PTY and its output is captured and cleaned up.

## Testing

The `./tests` subfolder in the source code contains test files that demonstrate how to use the `ptywrapper` module.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

ptywrapper is licensed under the MIT license.
