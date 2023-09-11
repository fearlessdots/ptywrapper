package main

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
