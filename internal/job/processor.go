package job

import "context"

type Processor interface {
	Process(ctx context.Context, job Job) error
}
