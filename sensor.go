package viamroscli

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
)

var ModelGenericSensor = family.WithModel("generic-sensor")

func init() {
	resource.RegisterComponent(
		sensor.API,
		ModelGenericSensor,
		resource.Registration[sensor.Sensor, *genericSensorConfig]{
			Constructor: newGenericSensor,
		})
}

type genericSensorConfig struct {
	RosRoot string `json:"ros_root"`
	Topic   string
}

func (cfg genericSensorConfig) Validate(path string) ([]string, error) {
	if cfg.Topic == "" {
		return nil, fmt.Errorf("need ropic")
	}
	return nil, nil
}

func newGenericSensor(ctx context.Context, deps resource.Dependencies, config resource.Config, logger logging.Logger) (sensor.Sensor, error) {
	newConf, err := resource.NativeConfig[*genericSensorConfig](config)
	if err != nil {
		return nil, err
	}

	s := &genericSensor{name: config.ResourceName(), config: newConf, logger: logger}
	err = s.start(ctx)
	if err != nil {
		return nil, err
	}

	return s, nil
}

type genericSensor struct {
	resource.AlwaysRebuild

	name   resource.Name
	logger logging.Logger
	config *genericSensorConfig

	lock      sync.Mutex
	lastValue map[string]interface{}
	lastError error

	out    chan []string
	wg     sync.WaitGroup
	cancel context.CancelFunc
}

func (cs *genericSensor) start(ctx context.Context) error {
	if cs.cancel != nil {
		return fmt.Errorf("already started")
	}

	ctx, cs.cancel = context.WithCancel(ctx)
	cs.out = make(chan []string)

	go cs.run(ctx)
	go cs.runReceiver(ctx)
	return nil
}

func (cs *genericSensor) runReceiver(ctx context.Context) {
	cs.wg.Add(1)
	defer cs.wg.Done()

	for {
		err := ctx.Err()
		if err != nil {
			return
		}

		select {
		case res := <-cs.out:
			msg, err := parseMessage(res)
			if err != nil {
				cs.logger.Errorf("error parsing message %v", err)
			}
			msg["_ts"] = time.Now()

			cs.lock.Lock()
			cs.lastValue = msg
			cs.lastError = err
			cs.lock.Unlock()

		case <-time.After(10 * time.Millisecond): // this is so we close quickly
			continue
		}

	}

}

func (cs *genericSensor) run(ctx context.Context) {
	cs.wg.Add(1)
	defer cs.wg.Done()

	for {
		err := ctx.Err()
		if err != nil {
			cs.logger.Infof("stopping genericSensor for topic (%s) because %v", cs.config.Topic, err)
			return
		}

		err = runRosTopic(ctx, cs.config.RosRoot, cs.config.Topic, cs.out, cs.logger)
		if err != nil {
			cs.logger.Warnf("got error running rostopic, sleeping and trying again %v", err)
		} else {
			cs.logger.Warnf("runRosTopic returned nothing, weird... sleeping and trying again")
		}

		time.Sleep(time.Second)
	}
}

func (cs *genericSensor) Readings(ctx context.Context, extra map[string]interface{}) (map[string]interface{}, error) {
	cs.lock.Lock()
	defer cs.lock.Unlock()
	return cs.lastValue, cs.lastError
}

func (cs *genericSensor) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	return nil, nil
}

func (cs *genericSensor) Close(ctx context.Context) error {
	if cs.cancel != nil {
		cs.cancel()
		close(cs.out)
	}
	cs.wg.Wait()
	return nil
}

func (cs *genericSensor) Name() resource.Name {
	return cs.name
}
