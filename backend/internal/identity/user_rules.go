package identity

import (
	"github.com/Aias00/cloudbase/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

func IsUserAdmin(role string) bool {
	return role == domain.RoleAdmin
}

func IsUserActive(status string) bool {
	return status == domain.StatusActive
}

func CanUserBindGroup(allowedGroups []int64, groupID int64, isExclusive bool) bool {
	if !isExclusive {
		return true
	}
	for _, id := range allowedGroups {
		if id == groupID {
			return true
		}
	}
	return false
}

func HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

func CheckPassword(password, hashedPassword string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) == nil
}
