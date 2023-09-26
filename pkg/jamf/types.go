package jamf

import "time"

type Jamf struct {
	ServerURL string
	AuthToken string
}

// serial number to prestage enrollment ID
type EnrollmentMapping map[string]int

// --- internal types for json deserializing ---

// /v2/mobile-devices/detail
type mobileDeviceDetail struct {
	MobileDeviceID string `json:"mobileDeviceId"`
	Hardware       struct {
		SerialNumber string `json:"serialNumber"`
	} `json:"hardware"`
	General struct {
		DeviceOwnershipType      string `json:"deviceOwnershipType"`
		EnrollmentMethodPrestage struct {
			MobileDevicePrestageID string `json:"mobileDevicePrestageId"`
			ProfileName            string `json:"profileName"`
		} `json:"enrollmentMethodPrestage"`
		LastEnrolledDate time.Time `json:"lastEnrolledDate"`
	} `json:"general"`
}

// continuation from above
type mobileDevicesDetailResult struct {
	TotalCount int                  `json:"totalCount"`
	Results    []mobileDeviceDetail `json:"results"`
}

// /v2/mobile-device-prestages/scope
type mobileDevicePrestageScopeResult struct {
	SerialsByPrestageID map[string]string `json:"serialsByPrestageId"`
}

// /v2/mobile-device-prestages
type mobileDevicePrestagesResult struct {
	TotalCount int `json:"totalCount"`
	Results    []struct {
		ID          string `json:"id"`
		DisplayName string `json:"displayName"`
		VersionLock int    `json:"versionLock"`
	} `json:"results"`
}

// prestage => version
type PrestageVersionLocks map[int]int
