package config

import (
	"encoding/json"
	"os"
)

// Config holds the bot configuration
type Config struct {
	Token               string   `json:"token"`
	AdminUserIDs        []string `json:"admin_user_ids"`
	AdminRoleIDs        []string `json:"admin_role_ids"`
	TriggerPhrases      map[string]string `json:"trigger_phrases"`
	BertifyExemptRoleID string   `json:"bertify_exempt_role_id"`
}

// Load reads the config file and returns the configuration
func Load() (*Config, error) {
	// Default config
	config := &Config{
		Token:        "",
		AdminUserIDs: []string{},
		AdminRoleIDs: []string{},
		TriggerPhrases: map[string]string{
			"hello":  "Hello there, I'm BotBert!",
			"hi":     "Hi there, friend!",
			"bert":   "Did someone call for a Bert? That's me!",
			"help":   "Need help? Try using /help or b!help for admin commands.",
		},
		BertifyExemptRoleID: "",
	}

	// Try to open config file
	file, err := os.Open("config.json")
	if err != nil {
		// If file doesn't exist, create it with default config
		if os.IsNotExist(err) {
			return createDefaultConfig(config)
		}
		return nil, err
	}
	defer file.Close()

	// Decode the file into the config struct
	decoder := json.NewDecoder(file)
	err = decoder.Decode(config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

// createDefaultConfig saves the default config to a file
func createDefaultConfig(config *Config) (*Config, error) {
	file, err := os.Create("config.json")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

// IsAdmin checks if a user has admin privileges
func IsAdmin(config *Config, userID string, memberRoles []string) bool {
	// Check if user ID is in admin list
	for _, adminID := range config.AdminUserIDs {
		if adminID == userID {
			return true
		}
	}

	// Check if user has admin role
	for _, userRole := range memberRoles {
		for _, adminRole := range config.AdminRoleIDs {
			if userRole == adminRole {
				return true
			}
		}
	}

	return false
}

// IsExemptFromBertify checks if a user is exempt from bertification
func IsExemptFromBertify(config *Config, memberRoles []string) bool {
	if config.BertifyExemptRoleID == "" {
		return false
	}

	for _, role := range memberRoles {
		if role == config.BertifyExemptRoleID {
			return true
		}
	}

	return false
}