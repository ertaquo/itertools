package itertools

type optionsSet struct {
	parallel      bool
	parallelLimit int
}

func toOptionsSet(options ...Option) optionsSet {
	var opts optionsSet
	for _, option := range options {
		option(&opts)
	}
	return opts
}

// Option configures an iterator operation that supports optional parallel execution.
type Option func(*optionsSet)

// WithParallel enables concurrent callback execution. Callbacks must be safe for
// concurrent use, and callback and result order are unspecified. The operation
// waits for all callbacks to finish, even if one returns an error. Concurrency is
// unlimited unless a positive limit is set with WithParallelLimit.
func WithParallel() Option {
	return func(o *optionsSet) {
		o.parallel = true
	}
}

// WithParallelLimit enables parallel execution with at most limit callbacks
// running at once. A nonpositive limit leaves concurrency unlimited. Other
// parallel behavior is the same as WithParallel.
func WithParallelLimit(limit int) Option {
	return func(o *optionsSet) {
		o.parallel = true
		o.parallelLimit = limit
	}
}
