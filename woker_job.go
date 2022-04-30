package gworker

import "context"

type WorkerJob interface {
	Run(ctx context.Context) error
}
