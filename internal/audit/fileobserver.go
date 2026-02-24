package audit

import (
	"encoding/json"
	"fmt"
	"os"
)

type FileAuditObserver struct {
	filePath string
	file     *os.File
}

func NewFileAuditObserver(filePath string) (*FileAuditObserver, error) {
	if filePath == "" {
		return nil, nil
	}

	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("open audit file: %w", err)
	}

	return &FileAuditObserver{
		filePath: filePath,
		file:     f,
	}, nil
}

func (f *FileAuditObserver) Notify(event AuditEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal audit event: %w", err)
	}

	data = append(data, '\n')

	_, err = f.file.Write(data)
	if err != nil {
		return fmt.Errorf("write audit file: %w", err)
	}

	return f.file.Sync()
}
