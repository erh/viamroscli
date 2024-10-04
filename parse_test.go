package viamroscli

import (
	"io"
	"os"
	"testing"
	"time"

	"go.viam.com/test"
)

func TestStream1(t *testing.T) {
	in, err := os.Open("testdata/stream1.txt")
	test.That(t, err, test.ShouldBeNil)

	c := make(chan []string, 100)
	num := 0
	
	go func() {
		for {
			m, more:= <- c
			if !more {
				return
			}
			test.That(t, len(m), test.ShouldEqual, 1)
			test.That(t, m[0], test.ShouldEqual, "data: False")
			num++
		}
	}()
	
	err = stream(in, c)
	test.That(t, err, test.ShouldEqual, io.EOF)

	for i := 0; i < 100; i++ {
		time.Sleep(time.Millisecond)
		if num >= 12 {
			break
		}
	}
	
	close(c)
	
	test.That(t, num, test.ShouldEqual, 12)
}

func TestStream2(t *testing.T) {
	in, err := os.Open("testdata/stream2.txt")
	test.That(t, err, test.ShouldBeNil)

	c := make(chan []string, 100)
	num := 0
	
	go func() {
		for {
			m, more:= <- c
			if !more {
				return
			}
			test.That(t, len(m), test.ShouldEqual, 2)
			test.That(t, m[0], test.ShouldEqual, "data: False")
			test.That(t, m[1], test.ShouldEqual, "x: 5")
			num++
		}
	}()
	
	err = stream(in, c)
	test.That(t, err, test.ShouldEqual, io.EOF)

	for i := 0; i < 100; i++ {
		time.Sleep(time.Millisecond)
		if num >= 2 {
			break
		}
	}
	
	close(c)
	
	test.That(t, num, test.ShouldEqual, 2)
}

func TestTryParseNumber(t *testing.T) {
	n, isNumber := tryParseNumber("123")
	test.That(t, isNumber, test.ShouldBeTrue)
	test.That(t, n, test.ShouldEqual, 123)

	n, isNumber = tryParseNumber("123.456")
	test.That(t, isNumber, test.ShouldBeTrue)
	test.That(t, n, test.ShouldEqual, 123.456)

	_, isNumber = tryParseNumber("1a23")
	test.That(t, isNumber, test.ShouldBeFalse)
}

func TestParseValue1(t *testing.T) {
	v, err := parseValue("False")
	test.That(t, err, test.ShouldBeNil)
	test.That(t, v, test.ShouldEqual, false)

	v, err = parseValue("True")
	test.That(t, err, test.ShouldBeNil)
	test.That(t, v, test.ShouldEqual, true)

	v, err = parseValue("\"4\"")
	test.That(t, err, test.ShouldBeNil)
	test.That(t, v, test.ShouldEqual, "4")

	v, err = parseValue("5")
	test.That(t, err, test.ShouldBeNil)
	test.That(t, v, test.ShouldEqual, 5)

	v, err = parseValue("[5,6]")
	test.That(t, err, test.ShouldBeNil)
	test.That(t, v, test.ShouldResemble, []interface{}{5,6})

}


func TestParse1(t *testing.T) {
	m, err := parseMessage([]string{ "data: False", "x: 5", "y:\"4\"" })
	test.That(t, err, test.ShouldBeNil)
	test.That(t, len(m), test.ShouldEqual, 3)
	test.That(t, m["data"], test.ShouldEqual, false)
	test.That(t, m["x"], test.ShouldEqual, 5)
	test.That(t, m["y"], test.ShouldEqual, "4")
}


func TestParseMsg1(t *testing.T) {
	in, err := os.Open("testdata/msg1.txt")
	test.That(t, err, test.ShouldBeNil)

	c := make(chan []string, 100)
	
	err = stream(in, c)
	test.That(t, err, test.ShouldEqual, io.EOF)

	lines := <- c

	close(c)

	test.That(t, len(lines), test.ShouldEqual, 8)
	
	msg, err := parseMessage(lines)
	test.That(t, err, test.ShouldBeNil)

	test.That(t, msg["format"], test.ShouldEqual, "jpeg")
	test.That(t, msg["data"], test.ShouldResemble, []interface{}{ 255, 216, 255, 219})

	m2, ok := msg["header"].(map[string]interface{})
	test.That(t, ok, test.ShouldBeTrue)
	test.That(t, m2["seq"], test.ShouldEqual, 5)
	test.That(t, m2["frame_id"], test.ShouldEqual, "image")

	m3, ok := m2["stamp"].(map[string]interface{})
	test.That(t, ok, test.ShouldBeTrue)
	test.That(t, m3["secs"], test.ShouldEqual, 1728049567)
	test.That(t, m3["nsecs"], test.ShouldEqual, 678641408)

	
}
