package triggers

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"botbert/config"

	"github.com/bwmarrin/discordgo"
)

// TriggersModule handles trigger words and responses
type TriggersModule struct {
	session *discordgo.Session
	config  *config.Config
}

// Name returns the module name
func (m *TriggersModule) Name() string {
	return "triggers"
}

// Initialize sets up the message triggers module
func (m *TriggersModule) Initialize(s *discordgo.Session, cfg *config.Config) error {
	m.session = s
	m.config = cfg
	return nil
}

// Cleanup performs any necessary cleanup
func (m *TriggersModule) Cleanup() error {
	fmt.Println("Cleaning up triggers module")
	return nil
}

// HandleMessageCreate is called whenever a message is created
func HandleMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Load config to get current trigger phrases
	cfg, err := config.Load()
	if err != nil {
		log.Printf("Failed to load config for message triggers: %v\n", err)
		return
	}

	// Convert message to lowercase for case-insensitive matching
	content := strings.ToLower(m.Content)

	// Check if the message contains any trigger phrases
	for trigger, response := range cfg.TriggerPhrases {
		if strings.Contains(content, strings.ToLower(trigger)) {
			fmt.Printf("Trigger '%s' matched in message '%s', sending response '%s'\n", trigger, m.Content, response)
			s.ChannelMessageSend(m.ChannelID, response)
			return
		}
	}
}

// AddTrigger adds a new trigger and response to the config
func AddTrigger(trigger, response string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// Add or update the trigger
	fmt.Printf("Adding trigger '%s' with response '%s'\n", trigger, response)
	cfg.TriggerPhrases[trigger] = response

	// Save the updated config
	file, err := createConfigFile()
	if err != nil {
		return err
	}
	defer file.Close()

	return saveConfig(file, cfg)
}

// RemoveTrigger removes a trigger from the config
func RemoveTrigger(trigger string) (bool, error) {
	cfg, err := config.Load()
	if err != nil {
		return false, err
	}

	// Check if the trigger exists
	_, exists := cfg.TriggerPhrases[trigger]
	if !exists {
		fmt.Printf("Trigger '%s' not found in config\n", trigger)
		return false, nil
	}

	// Remove the trigger
	fmt.Printf("Removing trigger '%s'\n", trigger)
	delete(cfg.TriggerPhrases, trigger)

	// Save the updated config
	file, err := createConfigFile()
	if err != nil {
		return true, err
	}
	defer file.Close()

	return true, saveConfig(file, cfg)
}

// ListTriggers returns all trigger phrases and their responses
func ListTriggers() (map[string]string, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	fmt.Printf("Listing triggers: %+v\n", cfg.TriggerPhrases)
	return cfg.TriggerPhrases, nil
}

// Helper functions for file operations - these would typically be in the config package
// but are duplicated here for simplicity in this example
func createConfigFile() (*os.File, error) {
	fmt.Println("Creating config file")
	return os.Create("config.json")
}

func saveConfig(file *os.File, cfg *config.Config) error {
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	fmt.Println("Saving config file")
	return encoder.Encode(cfg)
}
