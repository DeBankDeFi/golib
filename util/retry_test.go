package util_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/DeBankDeFi/golib/util"
)

func TestRetry(t *testing.T) {
	testCases := []util.RetryFunc{
		func() error {
			return nil
		},
		func() error {
			return util.RetryableError{}
		},
		func() error {
			return fmt.Errorf("hehe")
		},
	}

	for _, tc := range testCases {
		_ = util.Retry(tc, 10, time.Second)
	}
}
