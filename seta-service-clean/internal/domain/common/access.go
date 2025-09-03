package common

// Access represents the level of permission a user has on an asset.
type Access string

const (
	// ReadAccess allows viewing an asset.
	ReadAccess Access = "read"
	// WriteAccess allows modifying an asset.
	WriteAccess Access = "write"
)

// IsValid checks if the access level is either READ or WRITE.
func (a Access) IsValid() bool {
	switch a {
	case ReadAccess, WriteAccess:
		return true
	}
	return false
}
