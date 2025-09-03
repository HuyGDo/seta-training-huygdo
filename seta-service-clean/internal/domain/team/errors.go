package team

import "errors"

var (
	// ErrTeamMustHaveLead is returned when a team is created without a lead manager.
	ErrTeamMustHaveLead = errors.New("a team must have exactly one lead manager")
	// ErrCreatorNotManager is returned when the user creating the team is not in the initial list of managers.
	ErrCreatorNotManager = errors.New("the user creating the team must be a manager")
	// ErrMemberNotFound is returned when trying to remove a member that is not in the team.
	ErrMemberNotFound = errors.New("member not found in team")
)
