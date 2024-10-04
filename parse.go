package viamroscli

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"
)

func stream(ctx context.Context, rawIn io.Reader, msgHandler chan []string) error {
	in := bufio.NewReader(rawIn)

	lines := []string{}

	for {
		err := ctx.Err()
		if err != nil {
			return err
		}

		data, err := in.ReadString('\n')
		if err != nil {
			return err
		}

		data = strings.TrimRight(data, " \r\n")

		if data == "---" {
			msgHandler <- lines
			lines = []string{}
			continue
		}

		lines = append(lines, data)

	}

}

func tryParseNumber(s string) (interface{}, bool) {
	s = strings.TrimSpace(s)

	if len(s) == 0 {
		return nil, false
	}

	if !unicode.IsDigit(rune(s[0])) {
		return nil, false
	}

	hasDecimal := false
	for _, d := range s {
		if unicode.IsDigit(d) {
			continue
		}
		if d == '.' {
			if hasDecimal {
				return nil, false
			}
			hasDecimal = true
			continue
		}
		return nil, false
	}

	if hasDecimal {
		x, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return nil, false
		}
		return x, true
	}

	x, err := strconv.Atoi(s)
	if err != nil {
		return nil, false
	}
	return x, true
}

func parseValue(s string) (interface{}, error) {
	s = strings.TrimSpace(s)

	if len(s) == 0 {
		return nil, nil
	}

	if s == "False" {
		return false, nil
	}
	if s == "True" {
		return true, nil
	}

	if s[0] == '"' {
		if s[len(s)-1] != '"' {
			return nil, fmt.Errorf("invalid string (%s)", s)
		}
		return s[1 : len(s)-1], nil
	}

	if s[0] == '[' {
		x := s[1:]
		if s[len(s)-1] != ']' {
			return nil, fmt.Errorf("invalid array %v", s)
		}
		x = x[0 : len(x)-1]
		pcs := strings.Split(x, ",")

		arr := []interface{}{}

		for _, p := range pcs {
			v, err := parseValue(p)
			if err != nil {
				return nil, err
			}
			arr = append(arr, v)
		}

		return arr, nil
	}

	n, isNumber := tryParseNumber(s)
	if isNumber {
		return n, nil
	}

	return s, nil
}

type stack struct {
	m []map[string]interface{}
}

func (s *stack) pushNew(n string) {
	m := map[string]interface{}{}
	if len(n) > 0 {
		s.addToTop(n, m)
	}
	s.m = append(s.m, m)
}

func (s *stack) addToTop(n string, v interface{}) {
	m := s.m[len(s.m)-1]
	m[n] = v
}

func (s *stack) bottom() map[string]interface{} {
	return s.m[0]
}

func (s *stack) pop() {
	s.m = s.m[0 : len(s.m)-1]
}

func parseMessage(lines []string) (map[string]interface{}, error) {

	s := stack{}
	s.pushNew("")

	nextIndent := 0

	for _, l := range lines {
		split := strings.SplitN(l, ":", 2)
		if len(split) != 2 {
			return nil, fmt.Errorf("invalid line [%s]", l)
		}
		name := split[0]
		indent := len(name) // temp
		name = strings.TrimLeft(name, " ")
		indent = indent - len(name)

		if indent > nextIndent {
			return nil, fmt.Errorf("badly formatted message?? indent wrong")
		}

		if indent < nextIndent {
			s.pop()
			nextIndent = indent
		}

		name = strings.TrimSpace(name)

		rest := strings.TrimSpace(split[1])
		if rest == "" {
			s.pushNew(name)
			nextIndent += 2
			continue
		}
		v, err := parseValue(rest)
		if err != nil {
			return nil, fmt.Errorf("field: %s error parsing value (%s): %v", name, rest, err)
		}
		s.addToTop(name, v)
	}

	return s.bottom(), nil
}
