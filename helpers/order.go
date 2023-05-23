package helpers

import (
	"io"
	"strconv"
)

func LunaOrderCheck(orderID string) bool {
	if orderID == "" {
		return false
	}
	n := len(orderID)
	sum := 0
	second := false
	for i := n - 1; i >= 0; i-- {
		d, err := strconv.Atoi(string(orderID[i]))
		if err != nil {
			return false
		}

		if second {
			d = d * 2
		}

		sum += d/10 + d%10

		second = !second
	}

	return sum%10 == 0
}

func ReadBodyAsString(body io.ReadCloser) (string, error) {
	b, err := io.ReadAll(body)
	if err != nil {
		return "", err
	}

	s := string(b)
	return s, nil
}

func ReadBodyAsBytes(body io.ReadCloser) ([]byte, error) {
	b, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}

	return b, nil
}
