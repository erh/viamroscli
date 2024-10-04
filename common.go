package viamroscli

import (
	"context"
	"fmt"
	"os/exec"
	
	"go.viam.com/rdk/resource"
)

var family = resource.ModelNamespace("erh").WithFamily("viamroscli")


// runs in foreground, doesn't retry, etc.
// that should be done on top of this
func runRosTopic(ctx context.Context, root string, topic string, out chan[]string) error {
	cmd := "rostopic"
	if root != "" {
		cmd = fmt.Sprintf("%s/bin/rostopic", root)
	}
	
	c := exec.Command(cmd, "echo", topic)
	if root != "" {
		c.Env = append(c.Env, fmt.Sprintf("PYTHONPATH=%s/lib/python2.7/dist-packages", root)) // TODO - add more?
	}

	return runRosTopicExec(ctx, c, out)
}

func runRosTopicExec(ctx context.Context, c *exec.Cmd, out chan[]string) error {
	panic(1)
}
