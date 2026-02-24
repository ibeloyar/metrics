package audit

import (
	"fmt"
)

type AuditEvent struct {
	TS        int64    `json:"ts"` // timestamp
	Metrics   []string `json:"metrics"`
	IPAddress string   `json:"ip_address"`
}

type AuditObserver interface {
	Notify(event AuditEvent) error
}

type AuditSubject struct {
	observers []AuditObserver
}

func NewSubject() *AuditSubject {
	return &AuditSubject{}
}

func (s *AuditSubject) Register(obs AuditObserver) {
	if obs != nil {
		s.observers = append(s.observers, obs)
	}
}

func (s *AuditSubject) NotifyAll(event AuditEvent) {
	for _, obs := range s.observers {
		if err := obs.Notify(event); err != nil {
			fmt.Printf("audit observer error: %v\n", err)
		}
	}
}

func (f *FileAuditObserver) Close() error {
	if f.file != nil {
		return f.file.Close()
	}
	return nil
}
