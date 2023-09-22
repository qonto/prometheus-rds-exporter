package converter_test

import (
	"testing"

	converter "github.com/qonto/prometheus-rds-exporter/internal/app/unit"
	"github.com/stretchr/testify/assert"
)

func TestGigaBytesToBytes(t *testing.T) {
	assert.Equal(t, int64(1073741824), converter.GigaBytesToBytes(int32(1)), "1 GB conversion is not correct")
	assert.Equal(t, int64(1073741824), converter.GigaBytesToBytes(int64(1)), "1 GB conversion is not correct")
	assert.Equal(t, int64(1073741824), converter.GigaBytesToBytes(float64(1)), "1 GB conversion is not correct")
}

func TestMegaBytesToBytes(t *testing.T) {
	assert.Equal(t, int64(1048576), converter.MegaBytesToBytes(int64(1)), "1 MB conversion is not correct")
	assert.Equal(t, float64(1048576), converter.MegaBytesToBytes(float64(1)), "1 MB conversion is not correct")
}

func TestKiloByteToMegaBytes(t *testing.T) {
	assert.Equal(t, int64(1), converter.KiloByteToMegaBytes(int64(1024)), "1 MB conversion is not correct")
}

func TestDaystoSeconds(t *testing.T) {
	assert.Equal(t, int32(86400), converter.DaystoSeconds(int32(1)), "1 day conversion is not correct")
	assert.Equal(t, int32(604800), converter.DaystoSeconds(int32(7)), "7 days conversion is not correct")
}
