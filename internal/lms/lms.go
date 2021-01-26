// Package lms provides common interfaces for different authentication methods
// for a Learning Management System.
package lms

import (
	"strings"
)

// Service describes an LMS service.
type Service interface {
	Name() string
	Icon() string // URL
	Authorize() AuthorizationMethods
}

// IconSetter is an interface for services to optionally implement.
type IconSetter interface {
	SetIcon(url string)
}

// AuthorizationMethods contains possible authorization methods supported by a
// service's implementation. Fields that are not nil are supported.
type AuthorizationMethods struct {
	Token TokenAuthorization
}

// TokenAuthorization is an authorization method which uses direct tokens.
type TokenAuthorization interface {
	Authorize(token string) (Session, error)
}

// Session describes an authorized session.
type Session interface {
	// User returns the current user.
	User() (*User, error)
}

// User describes a user.
type User struct {
	ID     string
	Name   Name
	Avatar string // URL
}

// Name describes a name. If either First or Last is provided, then it is
// assumed that the field will contain the whole name (excluding Preferred).
type Name struct {
	First     string
	Last      string
	Preferred string
}

// String formats the name into one string.
func (name Name) String() string {
	if name.First == "" && name.Last == "" {
		return name.Preferred
	}

	builder := strings.Builder{}
	builder.Grow(len(name.First) + len(name.Last) + len(name.Preferred) + 10)

	switch {
	case name.First == "":
		builder.WriteString(name.Last)
	case name.Last == "":
		builder.WriteString(name.First)
	default:
		builder.WriteString(name.First)
		builder.WriteByte(' ')
		builder.WriteString(name.Last)
	}

	if name.Preferred != "" {
		builder.WriteByte(' ')
		builder.WriteByte('(')
		builder.WriteString(name.Preferred)
		builder.WriteByte(')')
	}

	return builder.String()
}
