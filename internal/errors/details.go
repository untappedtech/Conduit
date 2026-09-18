package errors

import "fmt"

func BuildDetails(err error, context ...string) []string {
	var details []string

	if err != nil {
		details = append(details, fmt.Sprintf("Error: %v", err))
	}

	for _, c := range context {
		if c != "" {
			details = append(details, c)
		}
	}

	return details
}
