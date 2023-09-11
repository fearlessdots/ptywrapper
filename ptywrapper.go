package ptywrapper

import (
  // Modules in GOROOT
  "os"
  "os/exec"
  "os/signal"
  "syscall"
  "sync"
  "time"
  "context"
  "bufio"
  "bytes"
  "strings"
  "regexp"
  "io"

  // External modules
  "github.com/creack/pty"
  "golang.org/x/term"
  "golang.org/x/sys/unix"
)

//
//// STRINGS
//

func cleanupString(originalString string) string {
  // Regular expression pattern to match ANSI escape sequences
  // This makes it easier to store, parse, read and use the command output as input for other programs
  reg := regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

  cleanedString := reg.ReplaceAllString(originalString, "")

  // Remove the first and last newline characters, if they exist.
  cleanedString = strings.TrimLeft(cleanedString, "\n")
  cleanedString = strings.TrimRight(cleanedString, "\n")

  // Remove all '\r' special characters
  cleanedString = strings.ReplaceAll(cleanedString, "\r", "")

  return cleanedString
}

//
//// CONTEXT
//


type contextWrapper struct {
  Ctx     context.Context
  Cancel  context.CancelFunc
}

func generateContextWrapper() contextWrapper {
  // Create a new context
	ctx, cancel := context.WithCancel(context.Background())

  // Generate a new contextWrapper
  wrapper := contextWrapper{
    Ctx: ctx,
    Cancel: cancel,
  }

  return wrapper
}

//
//// I/O WRITERS
//

type Writer struct {
  src     *os.File
  dest    *os.File
  ctx     contextWrapper
}

func (w *Writer) Write(p []byte) (n int, err error) {
  return w.dest.Write(p)
}

//
//// COMMAND
//

type Command struct {
  Entry       string
  Args        []string
  Env         []string
  Discard     bool // Discard output (will still save it as a variable)
  Completed   bool
  Output      string
  ExitCode    int
}

func (command *Command) RunInPTY() (Command, error) {
  // Create a command
  c := exec.Command(command.Entry, command.Args...)
  c.SysProcAttr = &syscall.SysProcAttr{
    Setctty: true, // Set controlling terminal to the pseudo terminal
    Setsid: true, // Start the command in a new session
  }

  // Set environment (use custom environment if available)
  if command.Env != nil {
    c.Env = command.Env
  } else {
    c.Env = os.Environ()
  }

  // Open a pty
  // Terminology:
  //   - primary => ptm (master)
  //   - secondary => pts (slave)
  primary, secondary, err := pty.Open()
  if err != nil {
    return *command, err
  }
  defer primary.Close()
  defer secondary.Close()

  // Set stdin, stdout and sterr for the command
  c.Stdin = secondary
  c.Stdout = secondary
  c.Stderr = secondary

  // Get the file descriptor for stdin
  fd := int(os.Stdin.Fd())

  // Make stdin raw and save the old state
  oldState, err := term.MakeRaw(fd)
  if err != nil {
    return *command, err
  }
  defer func() { _ = term.Restore(fd, oldState) }() // Ensure the old state is restored when the function returns

  // Enable non-blocking I/O on stdin
  flag, err := unix.FcntlInt(uintptr(fd), unix.F_GETFL, 0)
  if err != nil {
    return *command, err
  }
  flag, err = unix.FcntlInt(uintptr(fd), unix.F_SETFL, flag|unix.O_NONBLOCK)
  if err != nil {
    return *command, err
  }

  // Resize the pty
  ch := make(chan os.Signal, 1)
  errCh := make(chan error, 1)
  signal.Notify(ch, syscall.SIGWINCH)
  go func() {
    for range ch {
      if err := pty.InheritSize(os.Stdin, primary); err != nil {
        errCh <- err // Send the error to the error channel
        return
      }
    }
  }()
  ch <- syscall.SIGWINCH // Initial resize
  select {
  case err := <-errCh:
    return *command, err
  default:
    // No error, continue execution
  }
  func() { signal.Stop(ch); close(ch); close(errCh)}() // Cleanup signal and channels when done

  // Start the command
  err = c.Start()
  if err != nil {
    return *command, err
  }

  // Create a context to track if the command is still running
  cmdExecutionContext := generateContextWrapper()

  // Create a bytes buffer to capture the command's output
  var cmdOutput bytes.Buffer

  // Start goroutine to copy data from os.Stdin to ptm (via a custom writer)
  var ptyWriterWaitGroup sync.WaitGroup
  ptyWriterWaitGroup.Add(1)
  stdinWriter := &Writer{
    src: os.Stdin,
    dest: primary,
    ctx: cmdExecutionContext,
  }
  go func() {
    defer ptyWriterWaitGroup.Done()

    // Create a reader to get data from os.Stdin
    reader := bufio.NewReader(stdinWriter.src)

    // Create a bytes buffer
    buf := make([]byte, 4096)

    // Start loop
    Loop:
    for {
      select {
      case <-stdinWriter.ctx.Ctx.Done():
        break Loop
      default:
        // Read buffer
        n, err := reader.Read(buf)

        // Verify if there is no data available to read
        // NOTE: EAGAIN means that there is no data available to read, so it's not really an error in this case
        if pathErr, ok := err.(*os.PathError); ok && pathErr.Err == syscall.EAGAIN {
          // Sleep for 50 milliseconds (to slow down the loop but without removing too much fluidity from user input)
          time.Sleep(time.Millisecond * 10)

          continue
        } else {
          // Write data
          _, err = stdinWriter.Write(buf[:n])
        }
      }
    }

    return
  }()

  // Start goroutine to copy data from ptm to os.Stdout
  var stdoutWriterWaitGroup sync.WaitGroup
  stdoutWriterWaitGroup.Add(1)
  stdoutWriter := &Writer{
    src: primary,
    dest: os.Stdout,
    ctx: cmdExecutionContext,
  }
  go func() {
    defer stdoutWriterWaitGroup.Done()

    // Create a bytes buffer
    buf := make([]byte, 4096)

    // Start loop
    Loop:
    for {
      select {
      case <-stdoutWriter.ctx.Ctx.Done():
        break Loop
      default:
        // Read buffer
        n, err := stdoutWriter.src.Read(buf)

        // Verify if there is no data available to read
        // NOTE: EAGAIN means that there is no data available to read, so it's not really an error in this case
        if pathErr, ok := err.(*os.PathError); ok && pathErr.Err == syscall.EAGAIN {
          // Sleep for 50 milliseconds (to slow down the loop but without removing too much fluidity from the output)
          time.Sleep(time.Millisecond * 10)

          continue
        } else {
          // Write data
          if command.Discard == false {
            _, err = stdoutWriter.Write(buf[:n])
          }

          // Copy bytes to output bytes buffer
          _, err = io.Copy(&cmdOutput, bytes.NewReader(buf[:n]))
        }
      }
    }

    return
  }()

  // Wait for the command to exit
  cmdExitCh := make(chan error, 1)
  var cmdExecutionWaitGroup sync.WaitGroup
  cmdExecutionWaitGroup.Add(1)
  go func() {
    defer cmdExecutionWaitGroup.Done()

    cmdExitCh <- c.Wait()

    // Cancel context
    cmdExecutionContext.Cancel()

    // Close pty
    primary.Close()
    secondary.Close()

    // Wait for output writer to return
    stdoutWriterWaitGroup.Wait()

    // Wait for pty writer to return
    ptyWriterWaitGroup.Wait()

    return
  }()

  cmdExecutionWaitGroup.Wait()

  // Get command exit code and save it
  cmdExit := <-cmdExitCh
  close(cmdExitCh)
  if exitError, ok := cmdExit.(*exec.ExitError); ok {
    // The command exited with a non-zero status (an error)
    command.ExitCode = exitError.ExitCode()
  } else if cmdExit != nil {
    // Some other error occurred
    return *command, cmdExit
  } else {
    // The command exited successfully
    command.ExitCode = 0
  }

  // Convert command output from bytes to string
  cmdOutputString := cmdOutput.String()

  // Clean up command output
  cmdOutputStringCleaned := cleanupString(cmdOutputString)

  // Save cleaned up command output
  command.Output = cmdOutputStringCleaned

  // Mark command as completed and return
  command.Completed = true
  return *command, nil
}
