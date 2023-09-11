package main

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
