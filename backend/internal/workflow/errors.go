package workflow

import "errors"

var (
	// ErrUnknownNode is returned when a node code is not found in the definition.
	ErrUnknownNode = errors.New("unknown node code")

	// ErrInvalidTransition is returned when a transition between nodes is not allowed.
	ErrInvalidTransition = errors.New("invalid transition")

	// ErrTaskAlreadyCompleted is returned when the task has already been processed.
	ErrTaskAlreadyCompleted = errors.New("task already completed")

	// ErrInstanceNotRunning is returned when the process instance is not in running state.
	ErrInstanceNotRunning = errors.New("process instance is not running")

	// ErrCannotRejectFinalNode is returned when trying to reject at a non-rejectable node.
	ErrCannotRejectFinalNode = errors.New("this node does not support rejection")

	// ErrCannotRejectFromFirstNode is returned when trying to reject from the first node.
	ErrCannotRejectFromFirstNode = errors.New("cannot reject from the first node")

	// ErrVersionConflict is returned on optimistic lock conflict.
	ErrVersionConflict = errors.New("version conflict, retry later")
)