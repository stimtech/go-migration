package migration

import (
	"bytes"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_LoggerOption_apply(t *testing.T) {
	l := log.New(os.Stdout, "", 0)
	s := &Service{}
	LoggerOption{Logger: l}.apply(s)
	assert.Equal(t, l, s.logger.(defaultLogger).logger)
}

func Test_defaultLogger_Info(t *testing.T) {
	var buffer bytes.Buffer

	l := defaultLogger{logger: log.New(&buffer, "", 0)}

	l.Info("test message")
	assert.Equal(t, "test message\n", buffer.String())
}

func Test_defaultLogger_Warn(t *testing.T) {
	var buffer bytes.Buffer

	l := defaultLogger{logger: log.New(&buffer, "", 0)}

	l.Warn("test message")
	assert.Equal(t, "warning: test message\n", buffer.String())
}
