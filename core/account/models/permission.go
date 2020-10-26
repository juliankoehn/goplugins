package models

import (
	"fmt"
	"goplugins/core/framework"
)

type (
	// Group is a generic way of categorizing users to apply permissions, or
	// some other label, to those users. A user can belong to any number of
	// groups.
	Group struct {
		framework.Model
		Name        string        `json:"name"`
		Permissions []*Permission `json:"permissions" gorm:"many2many:groups_permissions;"`
	}
	// Permission system provides a way to assign permissions to specific
	// users and groups of users
	//
	// The Permission system is used by the AdminSite, but may also be useful in
	// your own code. The Admin size uses permissions as follows
	//
	//
	Permission struct {
		framework.Model
		Name        string   `json:"name"`
		ContentType string   `json:"contentType"`
		Codename    string   `json:"codeName"`
		Groups      []*Group `json:"groups" gorm:"many2many:groups_permissions;"`
	}
)

// String returns the Permission as String output
func (p *Permission) String() string {
	return fmt.Sprintf("%s | %s", p.Name, p.ContentType)
}

// String returns the Group as String
func (g *Group) String() string {
	return g.Name
}

// UserHasPerm validates the permission of a user
func UserHasPerm(user *User, perm string) bool {
	return false
}

// UserGetPermissions returns all permissions of given user
func UserGetPermissions(user *User, object string) []*Permission {
	return []*Permission{}
}
