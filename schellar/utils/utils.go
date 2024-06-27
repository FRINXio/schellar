package utils

import (
	"os"
	"strings"
)

func GetAdminValues() []string {

	adminRoles := []string{os.Getenv("ADMIN_ROLES"), os.Getenv("ADMIN_GROUPS")}

	// Filter out empty strings
	var nonEmptyRoles []string
	for _, header := range adminRoles {
		if header != "" {
			nonEmptyRoles = append(nonEmptyRoles, header)
		}
	}

	// Join non-empty headers with a comma
	adminRolesList := strings.Split(strings.Join(nonEmptyRoles, ","), ",")
	return RemoveDuplicates(adminRolesList)
}

// removeDuplicates removes duplicates from a slice of strings
func RemoveDuplicates(strings []string) []string {
	encountered := map[string]bool{} // Track encountered strings
	unique := []string{}             // Slice to hold unique strings

	for _, str := range strings {
		if !encountered[str] {
			encountered[str] = true
			unique = append(unique, str)
		}
	}

	return unique
}
