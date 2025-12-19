package slack

import (
	"encoding/json"
	"time"
)

// CredentialMetadata represents Slack-specific OAuth credential data
type CredentialMetadata struct {
	TeamID      string   `json:"team_id"`
	TeamName    string   `json:"team_name"`
	UserID      string   `json:"user_id,omitempty"`
	BotUserID   string   `json:"bot_user_id,omitempty"`
	Scope       string   `json:"scope"`
	AppID       string   `json:"app_id,omitempty"`
	TokenType   string   `json:"token_type"` // "bot" or "user"
}

// OAuthResponse represents the response from Slack OAuth2 token exchange
type OAuthResponse struct {
	OK          bool   `json:"ok"`
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
	BotUserID   string `json:"bot_user_id,omitempty"`
	AppID       string `json:"app_id"`
	Team        struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"team"`
	Enterprise         interface{} `json:"enterprise,omitempty"`
	AuthedUser         interface{} `json:"authed_user,omitempty"`
	IncomingWebhook    interface{} `json:"incoming_webhook,omitempty"`
	RefreshToken       string      `json:"refresh_token,omitempty"`
	ExpiresIn          int         `json:"expires_in,omitempty"`
	Error              string      `json:"error,omitempty"`
	ErrorDescription   string      `json:"error_description,omitempty"`
}

// APIResponse is the base Slack API response structure
type APIResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

// SendMessageRequest represents a request to send a message
type SendMessageRequest struct {
	Channel         string                   `json:"channel"`
	Text            string                   `json:"text,omitempty"`
	Blocks          []map[string]interface{} `json:"blocks,omitempty"`
	Attachments     []Attachment             `json:"attachments,omitempty"`
	ThreadTS        string                   `json:"thread_ts,omitempty"`
	ReplyBroadcast  bool                     `json:"reply_broadcast,omitempty"`
	UnfurlLinks     bool                     `json:"unfurl_links,omitempty"`
	UnfurlMedia     bool                     `json:"unfurl_media,omitempty"`
	IconEmoji       string                   `json:"icon_emoji,omitempty"`
	IconURL         string                   `json:"icon_url,omitempty"`
	Username        string                   `json:"username,omitempty"`
	AsUser          bool                     `json:"as_user,omitempty"`
	LinkNames       bool                     `json:"link_names,omitempty"`
	Mrkdwn          bool                     `json:"mrkdwn,omitempty"`
}

// Attachment represents a Slack message attachment (legacy)
type Attachment struct {
	Fallback   string                   `json:"fallback,omitempty"`
	Color      string                   `json:"color,omitempty"`
	Pretext    string                   `json:"pretext,omitempty"`
	AuthorName string                   `json:"author_name,omitempty"`
	AuthorLink string                   `json:"author_link,omitempty"`
	AuthorIcon string                   `json:"author_icon,omitempty"`
	Title      string                   `json:"title,omitempty"`
	TitleLink  string                   `json:"title_link,omitempty"`
	Text       string                   `json:"text,omitempty"`
	Fields     []AttachmentField        `json:"fields,omitempty"`
	ImageURL   string                   `json:"image_url,omitempty"`
	ThumbURL   string                   `json:"thumb_url,omitempty"`
	Footer     string                   `json:"footer,omitempty"`
	FooterIcon string                   `json:"footer_icon,omitempty"`
	Timestamp  int64                    `json:"ts,omitempty"`
	Actions    []map[string]interface{} `json:"actions,omitempty"`
}

// AttachmentField represents a field in an attachment
type AttachmentField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short,omitempty"`
}

// MessageResponse represents the response from sending/updating a message
type MessageResponse struct {
	OK      bool    `json:"ok"`
	Channel string  `json:"channel"`
	TS      string  `json:"ts"` // Message timestamp (unique ID)
	Message Message `json:"message,omitempty"`
	Error   string  `json:"error,omitempty"`
}

// Message represents a Slack message
type Message struct {
	Type        string                   `json:"type"`
	User        string                   `json:"user,omitempty"`
	Text        string                   `json:"text"`
	Blocks      []map[string]interface{} `json:"blocks,omitempty"`
	Attachments []Attachment             `json:"attachments,omitempty"`
	TS          string                   `json:"ts"`
	BotID       string                   `json:"bot_id,omitempty"`
	Username    string                   `json:"username,omitempty"`
	Icons       *MessageIcons            `json:"icons,omitempty"`
}

// MessageIcons represents message icons
type MessageIcons struct {
	Emoji       string `json:"emoji,omitempty"`
	Image48     string `json:"image_48,omitempty"`
	Image64     string `json:"image_64,omitempty"`
	Image72     string `json:"image_72,omitempty"`
}

// UpdateMessageRequest represents a request to update a message
type UpdateMessageRequest struct {
	Channel     string                   `json:"channel"`
	TS          string                   `json:"ts"`
	Text        string                   `json:"text,omitempty"`
	Blocks      []map[string]interface{} `json:"blocks,omitempty"`
	Attachments []Attachment             `json:"attachments,omitempty"`
	AsUser      bool                     `json:"as_user,omitempty"`
	LinkNames   bool                     `json:"link_names,omitempty"`
}

// User represents a Slack user
type User struct {
	ID                string      `json:"id"`
	TeamID            string      `json:"team_id"`
	Name              string      `json:"name"`
	Deleted           bool        `json:"deleted"`
	Color             string      `json:"color,omitempty"`
	RealName          string      `json:"real_name,omitempty"`
	TZ                string      `json:"tz,omitempty"`
	TZLabel           string      `json:"tz_label,omitempty"`
	TZOffset          int         `json:"tz_offset,omitempty"`
	Profile           UserProfile `json:"profile"`
	IsAdmin           bool        `json:"is_admin,omitempty"`
	IsOwner           bool        `json:"is_owner,omitempty"`
	IsPrimaryOwner    bool        `json:"is_primary_owner,omitempty"`
	IsRestricted      bool        `json:"is_restricted,omitempty"`
	IsUltraRestricted bool        `json:"is_ultra_restricted,omitempty"`
	IsBot             bool        `json:"is_bot,omitempty"`
	IsAppUser         bool        `json:"is_app_user,omitempty"`
	Updated           int64       `json:"updated,omitempty"`
}

// UserProfile represents a Slack user's profile
type UserProfile struct {
	FirstName          string `json:"first_name,omitempty"`
	LastName           string `json:"last_name,omitempty"`
	RealName           string `json:"real_name,omitempty"`
	RealNameNormalized string `json:"real_name_normalized,omitempty"`
	DisplayName        string `json:"display_name,omitempty"`
	Email              string `json:"email,omitempty"`
	Image24            string `json:"image_24,omitempty"`
	Image32            string `json:"image_32,omitempty"`
	Image48            string `json:"image_48,omitempty"`
	Image72            string `json:"image_72,omitempty"`
	Image192           string `json:"image_192,omitempty"`
	Image512           string `json:"image_512,omitempty"`
	StatusText         string `json:"status_text,omitempty"`
	StatusEmoji        string `json:"status_emoji,omitempty"`
	Team               string `json:"team,omitempty"`
}

// UserByEmailResponse represents the response from looking up a user by email
type UserByEmailResponse struct {
	OK    bool   `json:"ok"`
	User  User   `json:"user,omitempty"`
	Error string `json:"error,omitempty"`
}

// Conversation represents a Slack channel or DM
type Conversation struct {
	ID                 string   `json:"id"`
	Name               string   `json:"name,omitempty"`
	IsChannel          bool     `json:"is_channel"`
	IsGroup            bool     `json:"is_group"`
	IsIM               bool     `json:"is_im"`
	IsMpIM             bool     `json:"is_mpim"`
	IsPrivate          bool     `json:"is_private"`
	Created            int64    `json:"created"`
	IsArchived         bool     `json:"is_archived"`
	IsGeneral          bool     `json:"is_general,omitempty"`
	Unlinked           int      `json:"unlinked,omitempty"`
	NameNormalized     string   `json:"name_normalized,omitempty"`
	IsShared           bool     `json:"is_shared,omitempty"`
	IsOrgShared        bool     `json:"is_org_shared,omitempty"`
	IsMember           bool     `json:"is_member,omitempty"`
	IsReadOnly         bool     `json:"is_read_only,omitempty"`
	Creator            string   `json:"creator,omitempty"`
	Topic              *Topic   `json:"topic,omitempty"`
	Purpose            *Purpose `json:"purpose,omitempty"`
	NumMembers         int      `json:"num_members,omitempty"`
}

// Topic represents a channel topic
type Topic struct {
	Value   string `json:"value"`
	Creator string `json:"creator"`
	LastSet int64  `json:"last_set"`
}

// Purpose represents a channel purpose
type Purpose struct {
	Value   string `json:"value"`
	Creator string `json:"creator"`
	LastSet int64  `json:"last_set"`
}

// OpenConversationRequest represents a request to open a DM or group DM
type OpenConversationRequest struct {
	Users []string `json:"users"` // User IDs
}

// OpenConversationResponse represents the response from opening a conversation
type OpenConversationResponse struct {
	OK      bool         `json:"ok"`
	Channel Conversation `json:"channel,omitempty"`
	Error   string       `json:"error,omitempty"`
}

// ListChannelsResponse represents the response from listing channels
type ListChannelsResponse struct {
	OK       bool            `json:"ok"`
	Channels []*Conversation `json:"channels,omitempty"`
	Error    string          `json:"error,omitempty"`
	Metadata *Metadata       `json:"response_metadata,omitempty"`
}

// Metadata represents pagination metadata
type Metadata struct {
	NextCursor string `json:"next_cursor,omitempty"`
}

// ErrorResponse represents a Slack error response
type ErrorResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error"`
}

// ActionConfig represents the configuration for a Slack action in a workflow
type ActionConfig struct {
	// Common fields
	CredentialName string `json:"credential"` // References credential vault

	// Action-specific fields (will be different for each action type)
	Config json.RawMessage `json:"config"`
}

// SendMessageConfig represents configuration for SendMessage action
type SendMessageConfig struct {
	Channel         string                   `json:"channel"`
	Text            string                   `json:"text,omitempty"`
	Blocks          []map[string]interface{} `json:"blocks,omitempty"`
	ThreadTS        string                   `json:"thread_ts,omitempty"`
	ReplyBroadcast  bool                     `json:"reply_broadcast,omitempty"`
	UnfurlLinks     *bool                    `json:"unfurl_links,omitempty"`
	UnfurlMedia     *bool                    `json:"unfurl_media,omitempty"`
	IconEmoji       string                   `json:"icon_emoji,omitempty"`
	Username        string                   `json:"username,omitempty"`
}

// SendDMConfig represents configuration for SendDM action
type SendDMConfig struct {
	User   string                   `json:"user"` // User ID or email
	Text   string                   `json:"text,omitempty"`
	Blocks []map[string]interface{} `json:"blocks,omitempty"`
}

// UpdateMessageConfig represents configuration for UpdateMessage action
type UpdateMessageConfig struct {
	Channel string                   `json:"channel"`
	TS      string                   `json:"ts"`
	Text    string                   `json:"text,omitempty"`
	Blocks  []map[string]interface{} `json:"blocks,omitempty"`
}

// AddReactionConfig represents configuration for AddReaction action
type AddReactionConfig struct {
	Channel   string `json:"channel"`
	Timestamp string `json:"timestamp"`
	Emoji     string `json:"emoji"` // Emoji name without colons
}

// Validate validates SendMessageConfig
func (c *SendMessageConfig) Validate() error {
	if c.Channel == "" {
		return ErrChannelRequired
	}
	if c.Text == "" && len(c.Blocks) == 0 {
		return ErrTextOrBlocksRequired
	}
	if len(c.Text) > 40000 {
		return ErrTextTooLong
	}
	return nil
}

// Validate validates SendDMConfig
func (c *SendDMConfig) Validate() error {
	if c.User == "" {
		return ErrUserRequired
	}
	if c.Text == "" && len(c.Blocks) == 0 {
		return ErrTextOrBlocksRequired
	}
	if len(c.Text) > 40000 {
		return ErrTextTooLong
	}
	return nil
}

// Validate validates UpdateMessageConfig
func (c *UpdateMessageConfig) Validate() error {
	if c.Channel == "" {
		return ErrChannelRequired
	}
	if c.TS == "" {
		return ErrTimestampRequired
	}
	if c.Text == "" && len(c.Blocks) == 0 {
		return ErrTextOrBlocksRequired
	}
	if len(c.Text) > 40000 {
		return ErrTextTooLong
	}
	return nil
}

// Validate validates AddReactionConfig
func (c *AddReactionConfig) Validate() error {
	if c.Channel == "" {
		return ErrChannelRequired
	}
	if c.Timestamp == "" {
		return ErrTimestampRequired
	}
	if c.Emoji == "" {
		return ErrEmojiRequired
	}
	return nil
}

// UsageLog represents a Slack API usage log entry
type UsageLog struct {
	ID             string
	TenantID       string
	CredentialID   string
	ActionType     string
	ExecutionID    string
	Success        bool
	StatusCode     int
	ErrorMessage   string
	ResponseTimeMS int
	Metadata       map[string]interface{}
	CreatedAt      time.Time
}
