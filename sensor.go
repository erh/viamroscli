package viamroscli

import (
	"context"
	"fmt"
	"sync"

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
	err = s.start()
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
}

func (cs *genericSensor) start() error {
	panic(1)
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
	return nil
}

func (cs *genericSensor) Name() resource.Name {
	return cs.name
}
