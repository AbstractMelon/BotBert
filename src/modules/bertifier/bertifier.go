package bertifier

import (
	"fmt"
	"strings"

	"botbert/config"

	"github.com/bwmarrin/discordgo"
)

// BertifierModule handles the bertification of usernames
type BertifierModule struct {
	session *discordgo.Session
	config  *config.Config
}

// Name returns the module name
func (m *BertifierModule) Name() string {
	return "Bertifier"
}

// Initialize sets up the bertifier module
func (m *BertifierModule) Initialize(s *discordgo.Session, cfg *config.Config) error {
	fmt.Println("Initializing BertifierModule...")
	m.session = s
	m.config = cfg
	return nil
}

// Cleanup performs any necessary cleanup
func (m *BertifierModule) Cleanup() error {
	fmt.Println("Cleaning up BertifierModule...")
	return nil
}

// HandleGuildMemberAdd is called when a new member joins
func HandleGuildMemberAdd(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
	fmt.Printf("Handling new member join: %s\n", m.User.Username)
	// Check if member already has a nickname or if their username ends with "bert"
	if !needsBertification(m.User.Username, m.Nick) {
		fmt.Println("Member already bertified or does not need bertification.")
		return
	}

	// Check if member is exempt from bertification
	if config.IsExemptFromBertify(&config.Config{}, m.Roles) {
		fmt.Println("Member is exempt from bertification.")
		return
	}

	// Bertify the member
	BertifyMember(s, m.GuildID, m.User.ID, m.User.Username, m.Nick)
}

// BertifyMember adds "bert" to the end of a member's nickname if it doesn't already end with "bert"
func BertifyMember(s *discordgo.Session, guildID, userID, username, currentNick string) error {
    fmt.Printf("Bertifying member: %s\n", username)
    var newNick string

    // If the member has a current nickname
    if currentNick != "" {
        if !strings.HasSuffix(strings.ToLower(currentNick), "bert") {
            newNick = currentNick + "bert"
        } else {
            fmt.Println("Nickname already ends with 'bert'. No change needed.")
            return nil // Skip changing if already bertified
        }
    } else {
        // If no nickname, check the username
        if !strings.HasSuffix(strings.ToLower(username), "bert") {
            newNick = username + "bert"
        } else {
            fmt.Println("Username already ends with 'bert'. No change needed.")
            return nil // Skip changing if already bertified
        }
    }

    // Proceed with changing the nickname if needed
    fmt.Printf("Changing nickname to: %s\n", newNick)
    err := s.GuildMemberNickname(guildID, userID, newNick)
    if err != nil {
        fmt.Printf("Error changing nickname: %v\n", err)
    } else {
        fmt.Printf("Successfully changed nickname to: %s\n", newNick)
    }
    return err
}


// BertifyAllMembers is called when an admin uses the b!bertify command
func BertifyAllMembers(s *discordgo.Session, guildID string, cfg *config.Config) (int, error) {
	// Get all members in the guild
	var bertifiedCount int
	after := ""
	limit := 1000

	for {
		members, err := s.GuildMembers(guildID, after, limit)
		if err != nil {
			fmt.Println("Error retrieving guild members:", err)
			return bertifiedCount, err
		}

		if len(members) == 0 {
			break
		}

		// Process each member
		for _, member := range members {
			// Skip if member is exempt
			if config.IsExemptFromBertify(cfg, member.Roles) {
				fmt.Printf("Member %s is exempt from bertification.\n", member.User.Username)
				continue
			}

			// Check if member needs bertification
			if needsBertification(member.User.Username, member.Nick) {
				err := BertifyMember(s, guildID, member.User.ID, member.User.Username, member.Nick)
				if err == nil {
					bertifiedCount++
				}
			}
		}

		// Get the last member's ID for pagination
		after = members[len(members)-1].User.ID

		// If we got less than the limit, we're done
		if len(members) < limit {
			break
		}
	}

	fmt.Printf("Bertified %d members.\n", bertifiedCount)
	return bertifiedCount, nil
}

// needsBertification checks if a username needs to be bertified
func needsBertification(username, nickname string) bool {
	// If they have a nick, check if it ends with "bert"
	if nickname != "" {
		return !strings.HasSuffix(strings.ToLower(nickname), "bert")
	}
	
	// Otherwise, check if their username ends with "bert"
	return !strings.HasSuffix(strings.ToLower(username), "bert")
}

