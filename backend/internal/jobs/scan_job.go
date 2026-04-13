package jobs

import (
	"context"
	"log/slog"
)

const JobTypeScan = "scan"

type ScanPayload struct{}

type Scannable interface {
	Scan(ctx context.Context) error
}

func NewScanHandler(scanner Scannable, log *slog.Logger) Handler {
	return func(ctx context.Context, job *Job) error {
		return scanner.Scan(ctx)
	}
}
