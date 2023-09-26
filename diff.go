package main

import "github.com/umich-mac/jamf-prestage-fixup/pkg/jamf"

type diff struct {
	SerialNumber       string
	CurrentEnrollment  int
	PreviousEnrollment int
}

func threeDiff(previous jamf.EnrollmentMapping, current jamf.EnrollmentMapping) ([]diff, map[string]int, map[string]int) {
	notInCurrent := make(map[string]int)
	notInPrevious := make(map[string]int)

	var diffs []diff

	for serialNumber, previousPrestage := range previous {
		currentPrestage, found := current[serialNumber]
		if !found {
			notInCurrent[serialNumber] = previousPrestage
			continue
		}

		if previousPrestage != currentPrestage {
			diffs = append(diffs, diff{
				SerialNumber:       serialNumber,
				CurrentEnrollment:  currentPrestage,
				PreviousEnrollment: previousPrestage,
			})
		}

		// remove
		delete(current, serialNumber)
	}

	for serialNumber, ps := range current {
		notInPrevious[serialNumber] = ps
	}

	return diffs, notInPrevious, notInCurrent
}
