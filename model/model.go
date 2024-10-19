package model

// RiderRegistrationRequest represents the data required to register a rider
type RiderRegistrationRequest struct {
	PhoneNumber  string `json:"phone_number"`
	Password     string `json:"password"`
	Name         string `json:"name"`
	ProfileImage string `json:"profile_image"`
	LicensePlate string `json:"license_plate"`
}
