package goscope2

import "gorm.io/gorm"

type Config struct {
	DB *gorm.DB
	// default is 700
	LimitLogs int
	// Create a dictionary of the app id as key and app locations
	// e.g.:
	// ```
	// map[int][]string{239401: []string{"google.com"}}
	// ```
	AllowedApps map[int32][]string
	InternalApp int32
	AuthUser    string
	AuthPass    string
}

var Goscope2 struct{ config *Config }
