package viamroscli

import (
	"context"
	"io"
	"os/exec"
	"testing"
	"time"

	"go.viam.com/rdk/logging"
	"go.viam.com/test"
)

func TestCreateRosExec(t *testing.T) {
	c := createRosExec("/opt/ros/melodic", "foo")
	test.That(t, c.Path, test.ShouldEqual, "/opt/ros/melodic/bin/rostopic")
}

func TestRunRosTopic1(t *testing.T) {
	logger := logging.NewTestLogger(t)

	out := make(chan []string, 100)
	c := exec.Command("cat", "testdata/stream1.txt")
	err := runRosTopicExec(context.Background(), c, out, logger)
	test.That(t, err, test.ShouldEqual, io.EOF)

	count := 0
	done := false
	for !done {

		select {
		case res := <-out:
			msg, err := parseMessage(res)
			test.That(t, err, test.ShouldBeNil)
			count++
			test.That(t, msg, test.ShouldResemble, map[string]interface{}{"data": false})
		case <-time.After(10 * time.Millisecond):
			done = true
		}
	}

	test.That(t, count, test.ShouldEqual, 12)

}
