package install

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestValidatePasses(t *testing.T) {
	successValues := []DockerLogs{
		{MaxSize: "-1", MaxFile: 1},
		{MaxSize: "1k", MaxFile: 2},
		{MaxSize: "1m", MaxFile: 3},
		{MaxSize: "1g", MaxFile: 4},
	}
	for _, test := range successValues {
		valid, err := test.validate()

		assert.True(t, valid)
		assert.Empty(t, err)
	}
}

func TestValidateMaxFileFailsIfLessThanOne(t *testing.T) {
	logOpts := DockerLogs{MaxSize: "10m", MaxFile: 0}

	valid, err := logOpts.validate()

	assert.False(t, valid)
	assert.Len(t, err, 1)
	assert.Equal(t, err[0].Error(), "Max file must be greater than or equal to 1")
}

func TestValidateMaxSizeFailsIfOnlyWhitespace(t *testing.T) {
	logOpts := DockerLogs{MaxSize: " ", MaxFile: 1}

	valid, err := logOpts.validate()

	assert.False(t, valid)
	assert.Len(t, err, 1)
	assert.Equal(t, err[0].Error(), "Max size cannot be empty")
}

func TestValidateMaxSizeFailsIfEmpty(t *testing.T) {
	logOpts := DockerLogs{MaxSize: "", MaxFile: 1}

	valid, err := logOpts.validate()

	assert.False(t, valid)
	assert.Len(t, err, 1)
	assert.Equal(t, err[0].Error(), "Max size cannot be empty")
}

func TestValidateMaxSizeFailsIfMissingUnit(t *testing.T) {
	logOpts := DockerLogs{MaxSize: "10", MaxFile: 1}

	valid, err := logOpts.validate()

	assert.False(t, valid)
	assert.False(t, valid)
	assertErrorMsg(t, err, "Max size must be numberic followed by either k, m or g (lowercase)")
}

func TestValidateMaxSizeFailsIfInvalidValue(t *testing.T) {
	logOpts := DockerLogs{MaxSize: "invalidvalue", MaxFile: 1}

	valid, err := logOpts.validate()

	assert.False(t, valid)
	assertErrorMsg(t, err, "Max size must be numberic followed by either k, m or g (lowercase)")
}

func TestValidateMaxSizeFailsIfUppercaseUnit(t *testing.T) {
	failureValues := []DockerLogs{
		{MaxSize: "1K", MaxFile: 1},
		{MaxSize: "1M", MaxFile: 1},
		{MaxSize: "1G", MaxFile: 1},
	}
	for _, test := range failureValues {
		valid, err := test.validate()

		assert.False(t, valid)
		assertErrorMsg(t, err, "Max size must be numberic followed by either k, m or g (lowercase)")
	}
}

func assertErrorMsg(t *testing.T, err []error, msg string) {
	assert.Len(t, err, 1)
	assert.Equal(t, err[0].Error(), msg)
}
