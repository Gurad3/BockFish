package main

import (
	_ "net/http/pprof"
	"testing"
)

func TestMain(m *testing.M) {

	m.Run()

}
func TestMainFunc(t *testing.T) {
	// Simulate a test case
	benchsearch()
}
