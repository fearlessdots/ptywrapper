package main

import (
  // Modules in GOROOT
  "fmt"

  // Modules from the project
  "github.com/fearlessdots/ptywrapper"
)

func main() {
  fmt.Println("=> Testing module...")
  fmt.Println("")

  // Add commands here
  commands := []ptywrapper.Command{
    ptywrapper.Command{
      Entry:  "/usr/bin/bash",
      Args:   []string{},
    },
    ptywrapper.Command{
      Entry:  "/usr/bin/ipython",
      Args:   []string{},
    },
  }

  // Iterate through the array
  for _, command := range commands {
    fmt.Println("=========================================")
    fmt.Println("")
    fmt.Println(fmt.Sprintf("Running: %s", command.Entry))
    fmt.Println("")
    _, err := command.RunInPTY()
    if err != nil {
      panic(err)
    } else {
      fmt.Println("")
      fmt.Println("==> Finished")
    }
    fmt.Println("")
    fmt.Println("=========================================")
  }
}
