package audit

import (
	"bytes"
	"errors"
	"os"
	"testing"
)

type MockObserver struct {
	called bool
	event  AuditEvent
	err    error
}

func (m *MockObserver) Notify(event AuditEvent) error {
	m.called = true
	m.event = event
	return m.err
}

type MockFile struct {
	closed bool
	err    error
	writes []string
}

func (f *MockFile) Write(p []byte) (n int, err error) {
	f.writes = append(f.writes, string(p))
	return len(p), nil
}

func (f *MockFile) Close() error {
	f.closed = true
	return f.err
}

func TestNewSubject(t *testing.T) {
	subject := NewSubject()
	if subject == nil {
		t.Fatal("NewSubject should not return nil")
	}
	if len(subject.observers) != 0 {
		t.Fatal("NewSubject should return empty observers slice")
	}
}

func TestRegisterNilObserver(t *testing.T) {
	subject := NewSubject()
	initialLen := len(subject.observers)

	subject.Register(nil)

	if len(subject.observers) != initialLen {
		t.Fatal("Register should not add nil observer")
	}
}

func TestRegisterValidObserver(t *testing.T) {
	subject := NewSubject()
	mock := &MockObserver{}

	subject.Register(mock)

	if len(subject.observers) != 1 {
		t.Fatal("Register should add valid observer")
	}
}

func TestNotifyAllNoObservers(t *testing.T) {
	subject := NewSubject()

	// Should not panic
	subject.NotifyAll(AuditEvent{})
}

func TestNotifyAllSuccess(t *testing.T) {
	subject := NewSubject()
	mock := &MockObserver{}
	subject.Register(mock)

	event := AuditEvent{
		TS:        1234567890,
		Metrics:   []string{"metric1"},
		IPAddress: "192.168.1.1",
	}

	subject.NotifyAll(event)

	if !mock.called {
		t.Fatal("Observer Notify should be called")
	}
	if mock.event.TS != event.TS {
		t.Errorf("Expected event.TS %d, got %d", event.TS, mock.event.TS)
	}
	if mock.event.IPAddress != event.IPAddress {
		t.Errorf("Expected event.IPAddress %q, got %q", event.IPAddress, mock.event.IPAddress)
	}
}

func TestNotifyAllErrorLogged(t *testing.T) {
	// Capture stdout to verify error logging
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	subject := NewSubject()
	mock := &MockObserver{err: errors.New("test error")}
	subject.Register(mock)

	subject.NotifyAll(AuditEvent{})

	w.Close()
	os.Stdout = oldStdout

	output := make([]byte, 1024)
	n, _ := r.Read(output)
	if !bytes.Contains(output[:n], []byte("audit observer error")) {
		t.Fatal("Observer error should be logged to stdout")
	}
}

func TestFileAuditObserverCloseNilFile(t *testing.T) {
	f := &FileAuditObserver{}

	err := f.Close()

	if err != nil {
		t.Errorf("Close should return nil when file is nil, got %v", err)
	}
}
