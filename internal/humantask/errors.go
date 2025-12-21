package humantask

import "errors"

var (
	// ErrTaskNotFound is returned when a task is not found
	ErrTaskNotFound = errors.New("task not found")

	// ErrTaskNotPending is returned when trying to complete a non-pending task
	ErrTaskNotPending = errors.New("task is not in pending status")

	// ErrUnauthorized is returned when a user cannot complete a task
	ErrUnauthorized = errors.New("user not authorized to complete this task")

	// ErrInvalidTaskType is returned when task type is invalid
	ErrInvalidTaskType = errors.New("invalid task type")

	// ErrInvalidStatus is returned when status is invalid
	ErrInvalidStatus = errors.New("invalid task status")

	// ErrMissingRequiredField is returned when a required field is missing
	ErrMissingRequiredField = errors.New("missing required field")
)
