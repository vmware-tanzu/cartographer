// Copyright 2021 VMware
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package testing

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
