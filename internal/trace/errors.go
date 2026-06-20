package trace

import "errors"

var (
	ErrEmptyTrace      = errors.New("trace contains no spans")
	ErrMultipleRoots   = errors.New("trace has multiple root spans")
	ErrOrphanSpan      = errors.New("span has non-existent parent")
	ErrNoRootSpan      = errors.New("trace has no root span")
	ErrInvalidSpanID   = errors.New("invalid span ID")
)
