package util

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"
)

// CommandResp ...
type CommandResp struct {
	Output string
	Err    error
}

// StartCommand start a shell command.
func StartCommand(cmds string, envs []string) (<-chan *CommandResp, *exec.Cmd, error) {
	dirName, _ := os.Getwd()
	p := exec.Command("sh", "-c", cmds)
	p.Dir = dirName
	if envs != nil {
		p.Env = envs
	}
	p.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	ch := make(chan *CommandResp, 1)
	go func() {
		out, err := p.CombinedOutput()
		ch <- &CommandResp{
			Output: string(out),
			Err:    err,
		}
	}()
	return ch, p, nil
}

// ExecuteCommand ...
func ExecuteCommand(showOutput bool, cmd string, args ...string) (err error) {
	var output []byte
	defer func() {
		if showOutput {
			fmt.Println(string(output))
		}
	}()
	p := exec.Command(cmd, args...)
	output, err = p.CombinedOutput()
	if err != nil {
		return err
	}
	return
}

type result struct {
	out []byte
	err error
}

// ShExec ...
func ShExec(cmd string, seconds int) (string, error) {
	p := exec.Command("sh", "-c", cmd)
	outCh := make(chan result, 1)
	go func() {
		out, err := p.CombinedOutput()
		outCh <- result{
			out: out,
			err: err,
		}
	}()
	select {
	case out := <-outCh:
		return string(out.out), out.err
	case <-time.After(time.Second * time.Duration(seconds)):
		err := p.Process.Kill()
		if err != nil {
			return "", err
		}
		select {
		case out := <-outCh:
			return string(out.out), errors.New("exec timeout")
		case <-time.After(time.Second * 5):
			return "", errors.New("exec timeout")
		}
	}
}
