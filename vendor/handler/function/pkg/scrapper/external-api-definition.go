package scrapper

type MessageType string

const (
	Chat       MessageType = "general"
	Roll       MessageType = "rollresult"
	InlineRoll MessageType = "inlinerollresult"
	Whisper    MessageType = "whisper"
)

// swagger:model Message
// Message A Message as sent on the Roll20 chat
type Message struct {
	// Link to the player avatar
	Avatar string `json:"avatar"`
	// Sent timestamp. I don't know why it's called priority
	Priority float64 `json:".priority, string"`
	// No idea, not parsing
	// Signature string
	// Command having triggered the roll action. Ex 1d20
	OrigRoll string `json:"origRoll,omitempty"`
	// Either a sub JSON or an expression specifying the content to parse
	Content string `json:"content"`
	// Either a chat message, a roll or an inline roll, inline roll are sometimes regarded as chat message
	// for some reason
	Type MessageType `json:"type"`
	// Game specific ID of the player sending the message
	PlayerId string `json:"playerId"`
	// Character name of the player sending the message
	Who string `json:"who"`
}

// swagger:model Player
//Player A Roll20 Player as listed on the campaign page
type Player struct {
	// This player avatar URl. Can either be on roll20 CDN or external
	AvatarUrl string `json:"avatarUrl"`
	// Is the player a GM of the parsed game
	// There can be multiple GMs for a single game
	// required: true
	IsGm bool `json:"isGm"`
	// This player roll20 unique id
	// required: true
	// unique: true
	Roll20Id int `json:"roll20Id"`
	// This player username (not character name in game, roll20 username)
	// required: true
	Username string `json:"username"`
}

// A Roll20Account is a basic creds user in roll20
//
// As there is so such thing as a service account in Roll20,
// this has to be a real account
type Roll20Account struct {
	Login    string
	Password string
}

// Options All user-changeable options
type Options struct {
	// Should the bot account ignore itself when retrieving data ? Default : true
	IgnoreSelf bool
}

// MessageOptions All available options when fetching messages
type MessageOptions struct {
	// Include the dice rolls
	IncludeRolls bool
	// Include all the text messages
	IncludeChat bool
	// Include whispers
	IncludeWhispers bool
}

// Build a new Message Options, by default, roll and chat messages are include but whispers messages are ignored
func NewMessageOptions() *MessageOptions {
	return &MessageOptions{
		IncludeRolls:    true,
		IncludeChat:     true,
		IncludeWhispers: false,
	}
}

// Check whether the message should be kept with these options
func (options *MessageOptions) isAllowing(m Message) bool {
	isAllowedRoll := options.IncludeRolls && (m.Type == Roll || m.Type == InlineRoll)
	isAllowedChat := options.IncludeChat && m.Type == Chat
	isAllowedWhisper := options.IncludeWhispers && m.Type == Whisper

	return isAllowedRoll || isAllowedChat || isAllowedWhisper
}

type IncompleteError struct {
	Err error
}

func (r *IncompleteError) Error() string {
	return r.Err.Error()
}
