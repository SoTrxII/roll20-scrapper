package config_parser

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestDefinedKey(t *testing.T) {
	const TEST_KEY = "foo"
	const TEST_VALUE = "bar"
	os.Setenv(TEST_KEY, TEST_VALUE)
	env_array := []string{TEST_KEY}
	res, err := ParseEnv(env_array)
	assert.Nil(t, err)
	assert.Equal(t, TEST_VALUE, res[TEST_KEY])

}

func TestUndefinedKey(t *testing.T) {
	const TEST_KEY = "foo"
	os.Unsetenv(TEST_KEY)
	env_array := []string{TEST_KEY}
	_, err := ParseEnv(env_array)
	assert.Error(t, err)
}
