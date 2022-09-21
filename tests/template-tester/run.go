package template_tester

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"

	"github.com/hashicorp/go-multierror"
)

type cmd struct {
	argv   []string
	stdout *bytes.Buffer
	stderr *bytes.Buffer
}

func Cmd(argv ...string) *cmd {
	return &cmd{
		argv:   argv,
		stdout: new(bytes.Buffer),
		stderr: new(bytes.Buffer),
	}
}

func (c *cmd) Run(ctx context.Context) error {
	if err := c.run(ctx); err != nil {
		if c.stdout.Len() > 0 {
			err = multierror.Append(err, fmt.Errorf(
				"stdout: %s", c.stdout.String(),
			))
		}

		if c.stderr.Len() > 0 {
			err = multierror.Append(err, fmt.Errorf(
				"stderr: %s", c.stderr.String(),
			))
		}

		return err
	}

	return nil
}

func (c *cmd) RunWithOutput(ctx context.Context) (string, string, error) {
	if err := c.Run(ctx); err != nil {
		return "", "", err
	}

	return c.stdout.String(), c.stderr.String(), nil
}

func (c *cmd) run(ctx context.Context) error {
	command := exec.CommandContext(ctx, c.argv[0], c.argv[1:]...)
	command.Stdout = c.stdout
	command.Stderr = c.stderr

	errC := make(chan error)

	go func() {
		errC <- command.Run()
		close(errC)
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-errC:
			if err != nil {
				return fmt.Errorf("run: %w", err)
			}

			return nil
		}
	}
}
