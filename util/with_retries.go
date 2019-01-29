package util

import "time"

type RetryHandler = func() (interface{}, error)

func WithRetries(hdlr RetryHandler, retryCount int, retryInterval time.Duration) (interface{}, error) {
	var err error

	for i := 0; i < retryCount; i++ {
		res, err := hdlr()
		if err != nil {
			time.Sleep(retryInterval)
			continue
		}

		return res, nil
	}

	return nil, err
}
