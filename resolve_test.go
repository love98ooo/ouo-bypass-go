package ouo_bypass_go

import "testing"

func TestResolve(t *testing.T) {
	resolve, err := Resolve("https://ouo.press/Zu7Vs5")
	expected := "http://google.com"
	if err != nil {
		t.Errorf("Resolve() failed with error: %v", err)
	}
	if resolve != expected {
		t.Errorf("Resolve() failed, expected %v, got %v", expected, resolve)
	}
}
