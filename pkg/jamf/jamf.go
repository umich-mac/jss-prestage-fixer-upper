package jamf

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/carlmjohnson/requests"
	"golang.org/x/exp/slog"
)

func New(serverURL string, adminToken string) *Jamf {
	var jamf Jamf
	jamf.ServerURL = serverURL
	jamf.AuthToken = adminToken

	return &jamf
}

func (jamf *Jamf) GetDeviceLastEnrollments() (EnrollmentMapping, error) {
	devices, err := jamf.GetDevices()
	if err != nil {
		return nil, err
	}

	enrollments := make(EnrollmentMapping)
	for _, device := range devices {
		oldId, err := strconv.Atoi(device.General.EnrollmentMethodPrestage.MobileDevicePrestageID)

		// skip this if it's not an int (could be empty)
		if err != nil {
			continue
		}
		enrollments[device.Hardware.SerialNumber] = oldId
	}

	return enrollments, nil
}

func (jamf *Jamf) GetDevices() ([]mobileDeviceDetail, error) {
	// Call /api/v2/mobile-devices/detail over and over until we have them all
	var devices []mobileDeviceDetail

	// handle pagination
	page := 0
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		var result mobileDevicesDetailResult
		err := requests.
			URL(jamf.ServerURL).
			Path("/api/v2/mobile-devices/detail").
			Bearer(jamf.AuthToken).
			Param("section", "GENERAL", "HARDWARE").
			ParamInt("page", page).
			ParamInt("page-size", 500).
			ToJSON(&result).
			Fetch(ctx)

		if err != nil {
			return nil, err
		}

		devices = append(devices, result.Results...)
		slog.Debug("fetched devices from mobile-devices/detail", slog.Int("count", len(devices)))

		// did we get them all?
		if len(devices) >= result.TotalCount {
			break
		} else {
			page++
		}

		if page > 100 {
			return nil, fmt.Errorf("ran way off edge")
		}
	}

	return devices, nil
}

func (jamf *Jamf) GetCurrentPrestageScopings() (EnrollmentMapping, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var result mobileDevicePrestageScopeResult
	err := requests.
		URL(jamf.ServerURL).
		Path("/api/v2/mobile-device-prestages/scope").
		Bearer(jamf.AuthToken).
		ToJSON(&result).
		Fetch(ctx)

	if err != nil {
		return nil, err
	}

	slog.Debug("fetched prestage scopes from mobile-device-prestages/scope", slog.Int("count", len(result.SerialsByPrestageID)))

	// reshape
	mappings := make(EnrollmentMapping)
	for serialNumber, prestageString := range result.SerialsByPrestageID {
		prestageInt, err := strconv.Atoi(prestageString)

		// skip this if it's not an int (could be empty)
		if err != nil {
			continue
		}
		mappings[serialNumber] = prestageInt
	}

	return mappings, nil
}

func (jamf *Jamf) GetPrestageVersionLocks() (PrestageVersionLocks, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var result mobileDevicePrestagesResult
	err := requests.
		URL(jamf.ServerURL).
		Path("/api/v2/mobile-device-prestages").
		Bearer(jamf.AuthToken).
		ParamInt("page", 0).
		ParamInt("page-size", 500).
		ToJSON(&result).
		Fetch(ctx)

	if err != nil {
		return nil, err
	}

	slog.Debug("fetched prestages from mobile-device-prestages/scope", slog.Int("count", len(result.Results)))

	if result.TotalCount > len(result.Results) {
		return nil, fmt.Errorf("need to paginate, not implemented")
	}

	// reshape
	versions := make(PrestageVersionLocks)
	for _, prestage := range result.Results {
		prestageInt, err := strconv.Atoi(prestage.ID)
		if err != nil {
			continue
		}

		versions[prestageInt] = prestage.VersionLock
	}

	return versions, nil
}

// TODO consider if these two functions can be coalesced into a single function with a different URL path
func (jamf *Jamf) RemoveFromPrestage(prestageID int, versionLock int, serials []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	requestBody := struct {
		SerialNumbers []string `json:"serialNumbers"`
		VersionLock   int      `json:"versionLock"`
	}{
		VersionLock:   versionLock,
		SerialNumbers: serials,
	}

	var result struct {
		PrestageID  string `json:"prestageId"`
		Assignments []struct {
			SerialNumber string `json:"serialNumber"`
			UserAssigned string `json:"userAssigned"`
		} `json:"assignments"`
		VersionLock int `json:"versionLock"`
	}

	err := requests.
		URL(jamf.ServerURL).
		Pathf("/api/v2/mobile-device-prestages/%d/scope/delete-multiple", prestageID).
		Bearer(jamf.AuthToken).
		BodyJSON(requestBody).
		ToJSON(&result).
		Fetch(ctx)

	if err != nil {
		return err
	}

	return nil
}

func (jamf *Jamf) AddToPrestage(prestageID int, versionLock int, serials []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	requestBody := struct {
		SerialNumbers []string `json:"serialNumbers"`
		VersionLock   int      `json:"versionLock"`
	}{
		VersionLock:   versionLock,
		SerialNumbers: serials,
	}

	var result struct {
		PrestageID  string `json:"prestageId"`
		Assignments []struct {
			SerialNumber string `json:"serialNumber"`
			UserAssigned string `json:"userAssigned"`
		} `json:"assignments"`
		VersionLock int `json:"versionLock"`
	}

	err := requests.
		URL(jamf.ServerURL).
		Pathf("/api/v2/mobile-device-prestages/%d/scope", prestageID).
		Bearer(jamf.AuthToken).
		BodyJSON(requestBody).
		ToJSON(&result).
		Fetch(ctx)

	if err != nil {
		return err
	}

	return nil
}
