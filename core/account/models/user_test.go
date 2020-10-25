package models

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	testUserWithoutName = User{
		Email: "test@example.com",
	}
)

func TestUser(t *testing.T) {
	user := User{
		FirstName: "firstName",
		LastName:  "lastName",
		Email:     "testing@testify.tld",
	}

	fullname := user.GetFullName()
	require.Equal(t, fmt.Sprintf("%s %s", user.FirstName, user.LastName), fullname)

	// only firstName
	user.LastName = ""
	fullname = user.GetFullName()
	require.Equal(t, user.FirstName, fullname)

	// only lastName
	user.FirstName = ""
	user.LastName = "Griffin"
	fullname = user.GetFullName()
	require.Equal(t, user.LastName, fullname)

	// fallback to email
	user.LastName = ""
	fullname = user.GetFullName()
	require.Equal(t, user.Email, fullname)

	email := user.GetShortName()
	require.Equal(t, user.Email, email)
}
