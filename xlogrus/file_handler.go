package xlogrus

import (
	"github.com/pkg/errors"
	"io"
	"os"
)

// 日志分发

// to file

func NewFileHandler(fileName string) (io.WriteCloser, error) {
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return file, nil
}

// to kafka

type KafkaLog struct{}

func (k *KafkaLog) Write([]byte) (int, error) {
	// k.Send 调用kafka send逻辑
	return 0, nil
}
