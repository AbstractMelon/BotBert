package modules

import (
	"fmt"

	"botbert/config"
	"botbert/modules/bertifier"
	"botbert/modules/triggers"

	"github.com/bwmarrin/discordgo"
)

// ModuleManager handles all bot modules
type ModuleManager struct {
	Session *discordgo.Session
	Config  *config.Config
	Modules []Module
}

// Module interface that all modules must implement
type Module interface {
	Name() string
	Initialize(s *discordgo.Session, cfg *config.Config) error
	Cleanup() error
}

var Manager *ModuleManager

// Initialize sets up all modules
func Initialize(s *discordgo.Session, cfg *config.Config) {
	fmt.Println("Initializing modules...")

	Manager = &ModuleManager{
		Session: s,
		Config:  cfg,
		Modules: []Module{},
	}

	// Initialize all modules
	modules := []Module{
		&bertifier.BertifierModule{},
		&triggers.TriggersModule{},
	}

	for _, module := range modules {
		err := module.Initialize(s, cfg)
		if err != nil {
			// Log the error but continue with other modules
			fmt.Printf("Failed to initialize module %s: %v\n", module.Name(), err)
			continue
		}
		Manager.Modules = append(Manager.Modules, module)
		fmt.Printf("Initialized module %s\n", module.Name())
	}
}

// Cleanup performs cleanup for all modules
func Cleanup() {
	if Manager == nil {
		return
	}

	for _, module := range Manager.Modules {
		err := module.Cleanup()
		if err != nil {
			fmt.Printf("Failed to cleanup module %s: %v\n", module.Name(), err)
		}
	}
}

