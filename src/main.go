package main

import (
	"botbert/commands"
	"botbert/config"
	"botbert/modules"
	"botbert/modules/bertifier"
	triggers "botbert/modules/triggers"
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

// Bot parameters
var (
	GuildID        = flag.String("guild", "", "Test guild ID. If not set, commands will be registered globally")
	BotToken       = flag.String("token", "", "Bot token")
	RemoveCommands = flag.Bool("rmcmd", true, "Remove all commands after shutdowning or not")
)

var s *discordgo.Session

func init() {
	flag.Parse()
}

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Println("Error loading config:", err)
		return
	}

	// If token was not provided via flag, use from config
	if *BotToken == "" {
		*BotToken = cfg.Token
	}

	// Create a new Discord session
	s, err = discordgo.New("Bot " + *BotToken)
	if err != nil {
		fmt.Println("Error creating Discord session:", err)
		return
	}

	// Register handlers
	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		fmt.Printf("BotBert is online as: %v#%v\n", s.State.User.Username, s.State.User.Discriminator)
		err = s.UpdateGameStatus(0, "Blockbert's Adventures")
		if err != nil {
			fmt.Println("Error updating status:", err)
		}
	})

	// Initialize modules
	modules.Initialize(s, cfg)

	// Register the messageCreate handler for admin commands
	s.AddHandler(commands.HandleAdminCommands)

	// Register module-specific handlers
	s.AddHandler(bertifier.HandleGuildMemberAdd)
	s.AddHandler(triggers.HandleMessageCreate)

	// Open a websocket connection to Discord
	err = s.Open()
	if err != nil {
		fmt.Println("Error opening connection:", err)
		return
	}

	
	// Register slash commands
	fmt.Println("Registering slash commands...")
	commands.RegisterSlashCommands(s, *GuildID)
	fmt.Println("Slash commands registered.")

	fmt.Println("BotBert is now running. Press CTRL+C to exit.")

	// Create a context for shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling for graceful shutdown
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	// Start a goroutine that waits for the shutdown signal
	go func() {
		<-sc // Wait for shutdown signal
		cancel() // Cancel the context for immediate shutdown
	}()

	// Block the main goroutine until context is canceled
	<-ctx.Done()

	// Perform cleanup
	fmt.Println("Shutting down BotBert...")

	// Remove commands if requested
	if *RemoveCommands {
		commands.RemoveSlashCommands(s, *GuildID)
	}

	// Close Discord session
	s.Close()
}
