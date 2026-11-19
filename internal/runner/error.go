package runner

import "fmt"

type CommandError struct {
	Command string
	Err     error
}

func (e *CommandError) Error() string {
	return fmt.Sprintf("failed to run command. command = %s, err = %s", e.Command, e.Err)
}
