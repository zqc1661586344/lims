package workflow

import "errors"

var (
	ErrUnknownNode               = errors.New("unknown node code")
	ErrInvalidTransition         = errors.New("invalid transition")
	ErrTaskAlreadyCompleted      = errors.New("task already completed")
	ErrInstanceNotRunning        = errors.New("process instance is not running")
	ErrCannotRejectFinalNode     = errors.New("this node does not support rejection")
	ErrCannotRejectFromFirstNode = errors.New("cannot reject from the first node")
	ErrVersionConflict           = errors.New("version conflict, retry later")
	ErrDeptNotMatch              = errors.New("current user's department does not match the task's responsible department")
	ErrTaskAlreadyApproved       = errors.New("this node's business record can no longer be modified after workflow has advanced")
)
