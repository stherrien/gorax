package rbac

import "errors"

var (
	// ErrRoleNotFound is returned when a role is not found
	ErrRoleNotFound = errors.New("role not found")

	// ErrPermissionNotFound is returned when a permission is not found
	ErrPermissionNotFound = errors.New("permission not found")

	// ErrRoleAlreadyExists is returned when a role with the same name already exists
	ErrRoleAlreadyExists = errors.New("role already exists")

	// ErrSystemRoleCannotBeModified is returned when attempting to modify a system role
	ErrSystemRoleCannotBeModified = errors.New("system role cannot be modified")

	// ErrSystemRoleCannotBeDeleted is returned when attempting to delete a system role
	ErrSystemRoleCannotBeDeleted = errors.New("system role cannot be deleted")

	// ErrInvalidRoleName is returned when a role name is invalid
	ErrInvalidRoleName = errors.New("invalid role name")

	// ErrNoRolesProvided is returned when no roles are provided
	ErrNoRolesProvided = errors.New("no roles provided")

	// ErrPermissionDenied is returned when a user doesn't have required permission
	ErrPermissionDenied = errors.New("permission denied")

	// ErrUserNotFound is returned when a user is not found
	ErrUserNotFound = errors.New("user not found")
)
