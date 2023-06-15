package status

import (
	"testing"
)

func Test(t *testing.T) {
	status := New()
	t.Log(status.HTTPCode(), status.Name, status.Code, status.Error())
	if status.OK() {
		t.Log("OK")
	}
	status = NewInvalidArgument(123, "not found user")
	t.Log(status.HTTPCode(), status.Name, status.Code, status.Error())
	if status.OK() {
		t.Log("OK")
	}

	t.Log(123456789 / 1_000_000)
	t.Log((123456789 / 1_000) % 1_000)
	t.Log(123456789 % 1_000)
}
