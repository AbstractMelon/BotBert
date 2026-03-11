package commands

import (
	"fmt"
	"log"
	"strings"

	"botbert/config"
	"botbert/modules/bertifier"
	"botbert/modules/triggers"

	"github.com/bwmarrin/discordgo"
)

var (
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "help",
			Description: "Shows help information about BotBert",
		},
		{
			Name:        "triggers",
			Description: "List all message triggers",
		},
		{
			Name:        "ping",
			Description: "Checks if BotBert is online",
		},
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"help": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			helpText := "**BotBert Commands:**\n" +
				"/help - Shows this help message\n" +
				"/triggers - Lists all message triggers\n" +
				"/ping - Checks if BotBert is online\n\n" +
				"**Admin Commands:**\n" +
				"b!help - Shows admin help\n" +
				"b!bertify - Adds 'bert' to all members who don't already have it\n" +
				"b!addtrigger <trigger> <response> - Adds a new message trigger\n" +
				"b!removetrigger <trigger> - Removes a message trigger"

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: helpText,
				},
			})
			log.Printf("Handled help command from %s", i.Member.User.ID)
		},
		"triggers": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			triggers, err := triggers.ListTriggers()
			if err != nil {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Error listing triggers: " + err.Error(),
					},
				})
				log.Printf("Error listing triggers: %v", err)
				return
			}

			if len(triggers) == 0 {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "No triggers are currently set.",
					},
				})
				log.Printf("No triggers are currently set")
				return
			}

			var triggerList strings.Builder
			triggerList.WriteString("**Current Triggers:**\n")
			for trigger, response := range triggers {
				triggerList.WriteString(fmt.Sprintf("• \"%s\" → \"%s\"\n", trigger, response))
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: triggerList.String(),
				},
			})
			log.Printf("Listed %d triggers for %s", len(triggers), i.Member.User.ID)
		},
		"ping": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Pong! BotBert is alive and well!",
				},
			})
			log.Printf("Handled ping command from %s", i.Member.User.ID)
		},
	}
)

// RegisterSlashCommands registers all slash commands
func RegisterSlashCommands(s *discordgo.Session, guildID string) {
	for _, cmd := range commands {
		_, err := s.ApplicationCommandCreate(s.State.User.ID, guildID, cmd)
		if err != nil {
			log.Printf("Error creating command %s: %v\n", cmd.Name, err)
		}
	}

	// Add handler for interactions
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})
}

// RemoveSlashCommands removes all registered slash commands
func RemoveSlashCommands(s *discordgo.Session, guildID string) {
	// Get all registered commands
	commands, err := s.ApplicationCommands(s.State.User.ID, guildID)
	if err != nil {
		log.Printf("Error getting commands: %v\n", err)
		return
	}

	// Delete each command
	for _, cmd := range commands {
		err := s.ApplicationCommandDelete(s.State.User.ID, guildID, cmd.ID)
		if err != nil {
			log.Printf("Error deleting command %s: %v\n", cmd.Name, err)
		}
	}
}

// HandleAdminCommands processes b! prefixed admin commands
func HandleAdminCommands(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore messages from the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Check if message starts with the admin prefix
	if !strings.HasPrefix(m.Content, "b!") {
		return
	}

	// Split the command and arguments
	args := strings.Fields(m.Content)
	if len(args) == 0 {
		log.Printf("Ignoring empty command from %s", m.Author.Username)
		return
	}

	// Get the command without the prefix
	cmd := strings.ToLower(strings.TrimPrefix(args[0], "b!"))

	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Printf("Error loading config: %v\n", err)
		s.ChannelMessageSend(m.ChannelID, "Error loading config: "+err.Error())
		return
	}

	// Check if user has admin rights
	isAdmin := config.IsAdmin(cfg, m.Author.ID, m.Member.Roles)
	if !isAdmin {
		log.Printf("User %s does not have admin rights", m.Author.Username)
		s.ChannelMessageSend(m.ChannelID, "Sorry, you don't have permission to use admin commands.")
		return
	}

	// Process commands
	switch cmd {
	case "help":
		handleAdminHelp(s, m)
	case "bertify":
		handleBertify(s, m)
	case "addtrigger":
		handleAddTrigger(s, m)
	case "removetrigger":
		handleRemoveTrigger(s, m)
	default:
		log.Printf("Ignoring unknown command from %s: %s", m.Author.Username, cmd)
		s.ChannelMessageSend(m.ChannelID, "Unknown command. Use b!help for a list of commands.")
	}
}


// Admin command handlers
func handleAdminHelp(s *discordgo.Session, m *discordgo.MessageCreate) {
	helpText := "**BotBert Admin Commands:**\n" +
		"b!help - Shows this help message\n" +
		"b!bertify - Adds 'bert' to all members who don't already have it\n" +
		"b!addtrigger <trigger> <response> - Adds a new message trigger\n" +
		"b!removetrigger <trigger> - Removes a message trigger"

	s.ChannelMessageSend(m.ChannelID, helpText)
	log.Printf("Sent admin help to %s", m.Author.Username)
}

func handleBertify(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Send initial message
	msg, err := s.ChannelMessageSend(m.ChannelID, " Bertifying all members...")
	if err != nil {
		log.Printf("Error sending initial message: %v\n", err)
		return
	}

	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Printf("Error loading config: %v\n", err)
		s.ChannelMessageEdit(m.ChannelID, msg.ID, "Error loading config: "+err.Error())
		return
	}

	// Bertify all members
	count, err := bertifier.BertifyAllMembers(s, m.GuildID, cfg)
	if err != nil {
		log.Printf("Error bertifying members: %v\n", err)
		s.ChannelMessageEdit(m.ChannelID, msg.ID, "Error bertifying members: "+err.Error())
		return
	}

	// Update message with results
	s.ChannelMessageEdit(m.ChannelID, msg.ID, fmt.Sprintf("  Bertified %d members!", count))
	log.Printf("Bertified %d members for %s", count, m.Author.Username)
}

func handleAddTrigger(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Need at least 3 arguments: b!addtrigger <trigger> <response>
	args := strings.Fields(m.Content)
	if len(args) < 3 {
		log.Printf("Ignoring invalid addtrigger command from %s: %s", m.Author.Username, m.Content)
		s.ChannelMessageSend(m.ChannelID, "Usage: b!addtrigger <trigger> <response>")
		return
	}

	// Extract trigger and response
	trigger := args[1]
	response := strings.Join(args[2:], " ")

	// Add trigger
	err := triggers.AddTrigger(trigger, response)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error adding trigger: "+err.Error())
		return
	}

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("✅ Added trigger: \"%s\" → \"%s\"", trigger, response))
}

func handleRemoveTrigger(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Need exactly 2 arguments: b!removetrigger <trigger>
	args := strings.Fields(m.Content)
	if len(args) != 2 {
		s.ChannelMessageSend(m.ChannelID, "Usage: b!removetrigger <trigger>")
		return
	}

	// Extract trigger
	trigger := args[1]

	// Remove trigger
	removed, err := triggers.RemoveTrigger(trigger)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error removing trigger: "+err.Error())
		return
	}

	if removed {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("✅ Removed trigger: \"%s\"", trigger))
	} else {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("❌ Trigger \"%s\" not found", trigger))
	}
}

