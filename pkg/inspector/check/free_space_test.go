package check

import (
	"math"
	"testing"
)

func TestFreeSpaceSmallValue(t *testing.T) {
	c := FreeSpaceCheck{
		MinimumBytes: 1,
		Path:         "/",
	}
	ok, _ := c.Check()
	if !ok {
		t.Errorf("check returned true for a very small amount of free space")
	}
}

func TestFreeSpaceEntirelyTooLargeValue(t *testing.T) {
	c := FreeSpaceCheck{
		//note if you are running this test on AWS and they upgrade to zettabyte scale block stores, it will fail. sry my bad.
		MinimumBytes: uint64(math.Pow(1000, 7)),
		Path:         "/",
	}
	ok, _ := c.Check()
	if ok {
		t.Errorf("check returned true for a ludicrous amount of free space")
	}
}
