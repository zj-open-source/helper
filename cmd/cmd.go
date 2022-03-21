package cmd

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
)

type Cmd struct {
	execPath *string
	cmd      *exec.Cmd
}

func NewCmd(ctx context.Context, execPath *string) *Cmd {
	return &Cmd{
		execPath: execPath,
		cmd:      exec.CommandContext(ctx, *execPath),
	}
}

func (c *Cmd) AddArgs(args []string) *Cmd {
	c.cmd.Args = append(c.cmd.Args, args...)
	return c
}

func (c *Cmd) SyncExecute() ([]string, error) {
	cwd, _ := os.Getwd()
	fmt.Fprintf(os.Stdout, "%s %s\n", color.CyanString(path.Join(cwd, c.cmd.Dir)), strings.Join(c.cmd.Args, " "))

	var stdout, stderr []byte
	var errStdout, errStderr error
	stdoutIn, _ := c.cmd.StdoutPipe()
	stderrIn, _ := c.cmd.StderrPipe()
	err := c.cmd.Start()
	if err != nil {
		logrus.Errorf("cmd.Start() failed with '%s'", err.Error())
		return nil, err
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		stdout, errStdout = copyAndCapture(os.Stdout, stdoutIn)
		wg.Done()
	}()

	go func() {
		stderr, errStderr = copyAndCapture(os.Stderr, stderrIn)
		wg.Done()
	}()
	wg.Wait()

	err = c.cmd.Wait()
	if err != nil {
		if c.cmd.ProcessState.ExitCode() != 1 {
			logrus.Errorf("cmd.Run() failed with %s", err.Error())
			return nil, err
		}
	}

	if errStdout != nil {
		logrus.Error("failed to capture stdout")
		return nil, errStdout
	}

	if errStderr != nil {
		logrus.Error("failed to capture stderr")
		return nil, errStderr
	}

	return []string{string(stdout), string(stderr)}, nil
}

func (c *Cmd) AsyncExecute() error {
	cwd, _ := os.Getwd()
	fmt.Fprintf(os.Stdout, "%s %s\n", color.CyanString(path.Join(cwd, c.cmd.Dir)), strings.Join(c.cmd.Args, " "))
	{
		stdoutPipe, err := c.cmd.StdoutPipe()
		if err != nil {
			return err
		}
		go c.scanAndStdout(bufio.NewScanner(stdoutPipe))
	}
	{
		stderrPipe, err := c.cmd.StderrPipe()
		if err != nil {
			return err
		}
		go c.scanAndStderr(bufio.NewScanner(stderrPipe))
	}

	if err := c.cmd.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return err
	}
	return nil
}

func copyAndCapture(w io.Writer, r io.Reader) ([]byte, error) {
	var out []byte
	buf := make([]byte, 1024)
	for {
		n, err := r.Read(buf[:])
		if n > 0 {
			d := buf[:n]
			out = append(out, d...)
			_, err := w.Write(d)
			if err != nil {
				return out, err
			}
		}
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return out, err
		}
	}
}

func (c *Cmd) scanAndStdout(scanner *bufio.Scanner) {
	for scanner.Scan() {
		fmt.Fprintln(os.Stdout, scanner.Text())
	}
}

func (c *Cmd) scanAndStderr(scanner *bufio.Scanner) {
	for scanner.Scan() {
		regStr := `More than (\d+) frames duplicated`
		re, _ := regexp.Compile(regStr)
		dup := re.FindStringSubmatch(scanner.Text())
		if len(dup) >= 2 {
			ph, _ := strconv.ParseInt(dup[1], 10, 64)
			if ph >= 1000 {
				fmt.Println("Error：", scanner.Text())
				_ = c.cmd.Process.Kill()
				os.Stderr.WriteString(dup[0] + "  ")
				return
			}
		}
		fmt.Fprintln(os.Stderr, scanner.Text())
	}
}
