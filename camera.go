package viamroscli

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"sync"
	"time"

	"go.viam.com/rdk/components/camera"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
)

var ModelCamera = family.WithModel("camera")

func init() {
	resource.RegisterComponent(
		camera.API,
		ModelCamera,
		resource.Registration[camera.Camera, *rostopicConfig]{
			Constructor: newCamera,
		})
}

func newCamera(ctx context.Context, deps resource.Dependencies, config resource.Config, logger logging.Logger) (camera.Camera, error) {
	newConf, err := resource.NativeConfig[*rostopicConfig](config)
	if err != nil {
		return nil, err
	}

	s := &rosCamera{config: newConf, logger: logger}
	err = s.start()
	if err != nil {
		return nil, err
	}

	src, err := camera.NewVideoSourceFromReader(ctx, s, nil, camera.ColorStream)
	if err != nil {
		return nil, err
	}

	return camera.FromVideoSource(config.ResourceName(), src, logger), nil
}

type rosCamera struct {
	logger logging.Logger
	config *rostopicConfig

	lock      sync.Mutex
	lastValue image.Image
	lastError error

	out    chan []string
	wg     sync.WaitGroup
	cancel context.CancelFunc
}

func (rc *rosCamera) start() error {
	ctx := context.Background()
	if rc.cancel != nil {
		return fmt.Errorf("already started")
	}

	ctx, rc.cancel = context.WithCancel(ctx)
	rc.out = make(chan []string)

	go rc.run(ctx)
	go rc.runReceiver(ctx)
	return nil
}

func (rc *rosCamera) runReceiver(ctx context.Context) {
	rc.wg.Add(1)
	defer rc.wg.Done()

	for {
		err := ctx.Err()
		if err != nil {
			return
		}

		select {
		case res := <-rc.out:
			msg, err := parseMessage(res)
			var img image.Image

			if err != nil {
				rc.logger.Errorf("error parsing message %v", err)
			} else {
				img, err = getImage(msg)
				if err != nil {
					rc.logger.Errorf("error making image %v", err)
				}
			}

			rc.lock.Lock()
			rc.lastValue = img
			rc.lastError = err
			rc.lock.Unlock()

		case <-time.After(10 * time.Millisecond): // this is so we close quickly
			continue
		}

	}

}

func (rc *rosCamera) run(ctx context.Context) {
	rc.wg.Add(1)
	defer rc.wg.Done()

	for {
		err := ctx.Err()
		if err != nil {
			rc.logger.Infof("stopping rosCamera for topic (%s) because %v", rc.config.Topic, err)
			return
		}

		err = runRosTopic(ctx, rc.config.RosRoot, rc.config.Topic, rc.out, rc.logger)
		if err != nil {
			rc.logger.Warnf("got error running rostopic, sleeping and trying again %v", err)
		} else {
			rc.logger.Warnf("runRosTopic returned nothing, weird... sleeping and trying again")
		}

		time.Sleep(time.Second)
	}
}

func (rc *rosCamera) Read(ctx context.Context) (image.Image, func(), error) {
	rc.lock.Lock()
	defer rc.lock.Unlock()
	if rc.lastValue == nil && rc.lastError == nil {
		return nil, nil, fmt.Errorf("no image yet")
	}
	return rc.lastValue, nil, rc.lastError
}

func (rc *rosCamera) Close(ctx context.Context) error {
	if rc.cancel != nil {
		rc.cancel()
		close(rc.out)
	}
	rc.wg.Wait()
	return nil
}

func getImage(msg map[string]interface{}) (image.Image, error) {
	format, ok := msg["format"].(string)
	if !ok {
		return nil, fmt.Errorf("need a valid 'format' field %v", msg["format"])
	}

	data, ok := msg["data"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("need a valid 'data' array, got: %T", msg["data"])
	}

	realData := make([]byte, len(data))
	for idx, d := range data {
		i, ok := d.(int)
		if !ok {
			return nil, fmt.Errorf("array entry for image not an int, got %v %T", d, d)
		}
		if i < 0 || i > 255 {
			return nil, fmt.Errorf("array entry for image invalid, got %v", i)
		}
		realData[idx] = byte(i)
	}

	if format == "jpeg" {
		return jpeg.Decode(bytes.NewReader(realData))
	}

	return nil, fmt.Errorf("unknown format: [%s]", format)
}
