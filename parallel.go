package itertools

import "golang.org/x/sync/errgroup"

type ParallelOption func(wg *errgroup.Group)

func WithLimit(limit int) ParallelOption {
	return func(wg *errgroup.Group) {
		wg.SetLimit(limit)
	}
}
