package http_helpers

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

// This is only useful is the error object evolves
func TestFormatError(t *testing.T) {
	const MESSAGE = "meh"
	messageBytes := FormatError(MESSAGE)
	et := ErrorTemplate{}
	err := json.Unmarshal(messageBytes, &et)
	assert.Nil(t, err)
	assert.Equal(t, MESSAGE, et.Message)

}
