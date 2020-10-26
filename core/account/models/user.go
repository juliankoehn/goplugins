package models

import (
	"fmt"
	"goplugins/core/framework"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	// User reflects a user
	User struct {
		framework.Model
		ConfirmedAt        *time.Time    `json:"confirmedAt,omitempty" gorm:"column:confirmed_at"`
		ConfirmationToken  string        `json:"-" gorm:"column:confirmation_token"`
		ConfirmationSentAt *time.Time    `json:"confirmationSentAt,omitempty" gorm:"column:confirmation_sent_at"`
		Email              string        `json:"email" gorm:"uniqueIndex"`
		EmailChangeToken   string        `json:"-" gorm:"column:email_change_token"`
		EmailChange        string        `json:"newEmail,omitempty" gorm:"column:email_change"`
		EmailChangeSentAt  *time.Time    `json:"emailChangeSentAt,omitempty" gorm:"column:email_change_sent_at"`
		FirstName          string        `json:"firstName"`
		Groups             []*Group      `json:"groups" gorm:"many2many:user_groups;"` // The groups this user belongs to. A user will get all permissions granted to each of their groups
		InvitedAt          *time.Time    `json:"invitedAt,omitempty" gorm:"column:invited_at"`
		IsActive           bool          `json:"isActive"`     // Designates whether this user should be treated as active. Unselect this instead of deleting
		IsStaff            bool          `json:"isStaff"`      // Designates whether the user can log into this admin site.
		IsSuperUser        bool          `json:"isSuperUser"`  // Designates that this user has all permissions without explicitly assigning them.
		LastSignInAt       *time.Time    `json:"lastSignInAt"` // LastLogin is getting updated by the Store
		LastName           string        `json:"lastName"`
		Note               string        `json:"note"`
		Password           string        `json:"-"`
		Permissions        []*Permission `json:"permissions" gorm:"many2many:user_permissions;"` // Specific permissions for this user.
		RecoveryToken      string        `json:"-" gorm:"column:recovery_token"`
		RecoverySentAt     *time.Time    `json:"recoverySentAt,omitempty" gorm:"column:recovery_sent_at"`
		Username           string        `json:"username"`
	}
	// UserStore defines the user-repository
	UserStore interface {
		Create(user *User, password string) error
		SetEmail(user *User, email string) error
		Find(id uuid.UUID) (*User, error)
		FindByConfirmationToken(token string) (*User, error)
		FindByEmail(email string) (*User, error)
		FindByRecoveryToken(token string) (*User, error)
	}
)

// TableName returns the name of the database table
func (User) TableName() string {
	return "users"
}

// BeforeSave gets executed before the model is saved.
func (u *User) BeforeSave(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	if u.ConfirmedAt != nil && u.ConfirmedAt.IsZero() {
		u.ConfirmedAt = nil
	}

	if u.InvitedAt != nil && u.InvitedAt.IsZero() {
		u.InvitedAt = nil
	}
	if u.ConfirmationSentAt != nil && u.ConfirmationSentAt.IsZero() {
		u.ConfirmationSentAt = nil
	}
	if u.RecoverySentAt != nil && u.RecoverySentAt.IsZero() {
		u.RecoverySentAt = nil
	}
	if u.EmailChangeSentAt != nil && u.EmailChangeSentAt.IsZero() {
		u.EmailChangeSentAt = nil
	}
	if u.LastSignInAt != nil && u.LastSignInAt.IsZero() {
		u.LastSignInAt = nil
	}

	return nil
}

// IsConfirmed checks if a user is confirmed.
func (u *User) IsConfirmed() bool {
	return u.ConfirmedAt != nil
}

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
