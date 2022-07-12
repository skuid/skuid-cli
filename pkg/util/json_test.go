package util_test

import (
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/skuid/tides/pkg/util"
)

var (
	testData = []byte(`{"a":"b"}`)
)

type TestReadCloser struct {
	read  func(p []byte) (n int, err error)
	close func() error
}

func (trc TestReadCloser) Read(b []byte) (int, error) {
	return trc.read(b)
}

func (trc TestReadCloser) Close() error {
	return trc.close()
}

// TODO: finish this
func TestCombineJSON(t *testing.T) {
	var readIndex int64 = 0
	for _, tc := range []struct {
		description string
		readcloser  TestReadCloser
		reader      util.FileReader
		path        string

		expectedError error
	}{
		{
			description: "reader error",
			readcloser: TestReadCloser{
				read:  func(p []byte) (n int, err error) { return 0, fmt.Errorf("read") },
				close: func() error { return fmt.Errorf("close") },
			},
			reader: func(path string) ([]byte, error) {
				return []byte{}, fmt.Errorf("reader")
			},
			expectedError: fmt.Errorf("reader"),
		},
		{
			description: "read error",
			readcloser: TestReadCloser{
				read: func(p []byte) (n int, err error) { return len(p), fmt.Errorf("read") },
			},
			reader: func(path string) ([]byte, error) {
				return []byte{}, nil
			},
			expectedError: fmt.Errorf("read"),
		},
		{
			description: "invalid json patch",
			readcloser: TestReadCloser{
				read: func(p []byte) (n int, err error) { return 0, io.EOF },
			},
			reader: func(path string) ([]byte, error) {
				return []byte{}, nil
			},
			expectedError: fmt.Errorf("invalid json patch"),
		},
		{
			description: "success",
			readcloser: TestReadCloser{
				read: func(p []byte) (n int, err error) {
					if readIndex >= int64(len(testData)) {
						err = io.EOF
						return
					}

					n = copy(p, testData[readIndex:])
					readIndex += int64(n)
					return
				},
			},
			reader: func(path string) ([]byte, error) {
				return testData, nil
			},
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			_, err := util.CombineJSON(tc.readcloser, tc.reader, tc.path)
			if tc.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
