package main

import (
  // Modules in GOROOT
  "fmt"

  // Modules from the project
  "github.com/fearlessdots/ptywrapper"
)

func main() {
  // Add commands here
  commands := []ptywrapper.Command{
    ptywrapper.Command{
      Entry:  "/usr/bin/cat",
      Args:   []string{"./tests/test.json"},
      Discard: true,
    },
  }

  // Iterate through the array
  for _, command := range commands {
    response, err := command.RunInPTY()
    if err != nil {
      panic(err)
    }
    if response.ExitCode != 0 {
      fmt.Println(fmt.Sprintf("Exit code: %v", response.ExitCode))
      fmt.Println(fmt.Sprintf("An error has occurred: %s", response.Output))
    } else {
      fmt.Println(response.Output)
    }
  }
}
