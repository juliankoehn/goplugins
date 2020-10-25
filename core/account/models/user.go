package models

import (
	"fmt"
	"goplugins/core/framework"
	"strings"
	"time"
)

type (
	// User reflects a user
	User struct {
		framework.Model
		Email       string        `json:"email"`
		FirstName   string        `json:"firstName"`
		Groups      []*Group      `json:"groups"`      // The groups this user belongs to. A user will get all permissions granted to each of their groups
		IsActive    bool          `json:"isActive"`    // Designates whether this user should be treated as active. Unselect this instead of deleting
		IsStaff     bool          `json:"isStaff"`     // Designates whether the user can log into this admin site.
		IsSuperUser bool          `json:"isSuperUser"` // Designates that this user has all permissions without explicitly assigning them.
		LastLogin   time.Time     `json:"lastLogin"`   // LastLogin is getting updated by the Store
		LastName    string        `json:"lastName"`
		Note        string        `json:"note"`
		Password    string        `json:"password"`
		Permissions []*Permission `json:"permissions"` // Specific permissions for this user.
		Username    string        `json:"username"`
	}
	// Store defines the user-repository
	UserStore interface{}
)

// GetFullName returns the FullName of the User
func (u *User) GetFullName() string {
	if u.FirstName != "" || u.LastName != "" {
		return strings.Trim(fmt.Sprintf("%s %s", u.FirstName, u.LastName), " ")
	}

	return u.Email
}

// GetShortName returns the E-Mail of the User
func (u *User) GetShortName() string {
	return u.Email
}

// GetPermissions return a list of permissions
// that the user has directly.
func (u *User) GetPermissions() []*Permission {
	return UserGetPermissions(u, "user")
}

// GetGroupPermissions returns a list of permissions
// that this user has through their groups.
func (u *User) GetGroupPermissions() []*Permission {
	return UserGetPermissions(u, "group")
}

// GetAllPermissions returns all Permissions of the User
//
// This includes Grouped and User permissions
func (u *User) GetAllPermissions() []*Permission {
	return UserGetPermissions(u, "")
}

// HasPerm checks if the user has reqeusted Permission
// u.HasPerm("product-list")
//
// Returns true if the user has the specified permission.
func (u *User) HasPerm(perm string) bool {
	if u.IsActive && u.IsSuperUser {
		return true
	}
	return UserHasPerm(u, perm)
}
