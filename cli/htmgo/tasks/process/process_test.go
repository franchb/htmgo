package process

import (
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	if os.Getenv("HTMGO_TEST_SLEEPER") == "1" {
		time.Sleep(time.Minute)
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func startSleeper(t *testing.T, name string) <-chan error {
	t.Helper()
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("HTMGO_TEST_SLEEPER", "1")
	done := make(chan error, 1)
	go func() { done <- Run(NewRawCommand(name, exe, Silent)) }()
	deadline := time.Now().Add(10 * time.Second)
	for GetProcessByName(name) == nil {
		if time.Now().After(deadline) {
			t.Fatal("sleeper did not start")
		}
		time.Sleep(10 * time.Millisecond)
	}
	return done
}

func TestStop_KillsAndWaitsForExit(t *testing.T) {
	done := startSleeper(t, "test-sleeper")
	pid := int32(GetProcessByName("test-sleeper").Cmd.Process.Pid)

	Stop("test-sleeper", 5*time.Second)

	if PidExists(pid) {
		t.Fatal("process still running after Stop")
	}
	select {
	case err := <-done:
		if err != nil { // a deliberate stop is not an error, on any OS
			t.Fatalf("Run returned %v after a deliberate stop", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return after Stop")
	}
	if GetProcessByName("test-sleeper") != nil {
		t.Fatal("stopped process is still registered")
	}
}

func TestStop_UnknownNameIsNoop(t *testing.T) {
	Stop("not-running", time.Second)
}
