package aggregator

import (
	"bufio"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"time"
)

func (a *Aggregator) Commit() error {

	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	err = ExecCommand(dir, "git", "add", "docs/")
	if err != nil {
		return err
	}

	loc, err := time.LoadLocation("America/Denver")
	if err != nil {
		return err
	}
	t := time.Now().In(loc).Format(time.RFC850)
	err = ExecCommand(dir, "git", "commit", "-m", fmt.Sprintf("content update %s", t))
	if err != nil {
		return err
	}

	err = ExecCommand(dir, "git", "push", "origin", "master")
	if err != nil {
		return err
	}

	return nil
}

func ExecCommand(dir string, command string, arg ...string) error {

	cmd := exec.Command(command, arg...)
	cmd.Dir = dir

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("could not create standard error pipe, %s", err.Error())
	}
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			log.Info(scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			log.WithError(err).Info("failed to read from process standard error")
		}
	}()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("could not create standard out pipe, %s", err.Error())
	}
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			log.Info(scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			log.WithError(err).Info("failed to read from process standard out")
		}
	}()

	return cmd.Run()
}