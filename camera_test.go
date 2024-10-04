package viamroscli

import (
	"image/jpeg"
	"os"
	"testing"

	"go.viam.com/test"
)

func TestCam1(t *testing.T) {
	msg, err := read1MessageFromAFile("testdata/pic1.txt")
	test.That(t, err, test.ShouldBeNil)
	test.That(t, msg["format"], test.ShouldEqual, "jpeg")

	img, err := getImage(msg)
	test.That(t, err, test.ShouldBeNil)

	f, err := os.Create("/tmp/pic1.jpg")
	test.That(t, err, test.ShouldBeNil)
	defer f.Close()

	err = jpeg.Encode(f, img, nil)
	test.That(t, err, test.ShouldBeNil)
}
