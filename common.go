package viamroscli

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"

	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
)

var family = resource.ModelNamespace("erh").WithFamily("viamroscli")

// runs in foreground, doesn't retry, etc.
// that should be done on top of this
func runRosTopic(ctx context.Context, root string, topic string, out chan []string, logger logging.Logger) error {
	return runRosTopicExec(ctx, createRosExec(root, topic), out, logger)
}

func createRosExec(root string, topic string) *exec.Cmd {

	cmd := "rostopic"
	if root != "" {
		cmd = fmt.Sprintf("%s/bin/rostopic", root)
	}

	c := exec.Command(cmd, "echo", topic)
	if root != "" {
		c.Env = append(c.Env, fmt.Sprintf("PYTHONPATH=%s/lib/python2.7/dist-packages", root)) // TODO - add more?
	}

	return c
}

func runRosTopicExec(ctx context.Context, c *exec.Cmd, out chan []string, logger logging.Logger) error {
	stderr, err := c.StderrPipe()
	if err != nil {
		return err
	}

	stdout, err := c.StdoutPipe()
	if err != nil {
		return err
	}

	err = c.Start()
	if err != nil {
		return err
	}

	go func() {

		x := bufio.NewReader(stderr)

		for {
			l, err := x.ReadString('\n')
			if err != nil {
				return
			}
			if l != "" {
				logger.Errorf("stderr from rostopic: %s", l)
			}
		}
	}()

	err = stream(ctx, stdout, out)
	if err != nil {
		return err
	}

	return c.Wait()
}
