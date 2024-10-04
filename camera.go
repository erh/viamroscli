package viamroscli

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
)

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
