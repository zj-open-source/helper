package cmd

import (
	"context"
	"fmt"
	"testing"
)

func TestNewCmd(t *testing.T) {
	t.Run("CMD pwd use", func(t *testing.T) {
		command := "ps"
		cmd := NewCmd(context.Background(), &command)
		out, err := cmd.SyncExecute()
		if err != nil {
			return
		}

		fmt.Println("0 == ", out[0])
		fmt.Println("1 == ", out[1])
	})
	t.Run("CMD 'ps -ef' use", func(t *testing.T) {
		command := "ps"
		cmd := NewCmd(context.Background(), &command)
		cmd.AddArgs([]string{"-ef"})

		out, err := cmd.SyncExecute()
		if err != nil {
			return
		}

		fmt.Println("0 == ", out[0])
		fmt.Println("1 == ", out[1])
	})
}
