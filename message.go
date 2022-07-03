package guildrone

import "time"

const (
	MessageTypeDefault MessageType = "default"
	MessageTypeSystem  MessageType = "system"
)

// ChatMessage stores the data of a chat message
type ChatMessage struct {
	// The ID of the message
	ID string `json:"id"`

	// Type of the message
	Type MessageType `json:"type"`

	// The ID of the server the message was sent in
	ServerID string `json:"serverId"`

	// The ID of the channel the message was sent in
	ChannelID string `json:"channelId"`

	// Content of the message
	Content string `json:"content"`

	// Embedded content
	Embeds []ChatEmbed `json:"embeds"`

	// Message ids that were replied to
	ReplyMessageIds []string `json:"replyMessageIds"`

	// If set, this message will only be seen by those mentioned or replied to
	IsPrivate          bool       `json:"isPrivate"`
	IsSilent           bool       `json:"isSilent"`
	Mentions           *Mentions  `json:"mentions,omitempty"`
	CreatedAt          string     `json:"createdAt"`
	CreatedBy          string     `json:"createdBy"`
	CreatedByWebhookId string     `json:"createdByWebhookId"`
	UpdatedAt          *Timestamp `json:"updatedAt,omitempty"`
}

type MessageType string

type ChatEmbed struct {
	Title       string              `json:"title,omitempty"`
	Description string              `json:"description,omitempty"`
	URL         string              `json:"url,omitempty"`
	Color       int                 `json:"color"`
	Footer      ChatEmbedFooter     `json:"footer"`
	Timestamp   *Timestamp          `json:"timestamp,omitempty"`
	Thumbnail   *ChatEmbedThumbnail `json:"thumbnail,omitempty"`
	Image       *ChatEmbedImage     `json:"image,omitempty"`
	Author      *ChatEmbedAuthor    `json:"author,omitempty"`
	Fields      []ChatEmbedField    `json:"fields,omitempty"`
}

type ChatEmbedFooter struct {
	Text    string `json:"text,omitempty"`
	IconUrl string `json:"icon_url,omitempty"`
}

type ChatEmbedThumbnail struct {
	URL string `json:"url,omitempty"`
}

type ChatEmbedImage struct {
	URL string `json:"url,omitempty"`
}

type ChatEmbedAuthor struct {
	Name    string `json:"name,omitempty"`
	URL     string `json:"url,omitempty"`
	IconUrl string `json:"icon_url,omitempty"`
}

type ChatEmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

type Mentions struct {
	Users    []MentionUser    `json:"users,omitempty"`
	Channels []MentionChannel `json:"channels,omitempty"`
	Roles    []MentionRole    `json:"roles,omitempty"`
	Everyone bool             `json:"everyone,omitempty"`
	Here     bool             `json:"here,omitempty"`
}

type MentionUser struct {
	ID string `json:"id"`
}

type MentionChannel struct {
	ID string `json:"id"`
}

type MentionRole struct {
	ID string `json:"id"`
}

type MessageCreate struct {
	IsPrivate       bool        `json:"isPrivate,omitempty"`
	IsSilent        bool        `json:"isSilent,omitempty"`
	ReplyMessageIds []string    `json:"replyMessageIds,omitempty"`
	Content         string      `json:"content,omitempty"`
	Embeds          []ChatEmbed `json:"embeds,omitempty"`
}

type MessagesRequest struct {
	Before         *time.Time `json:"before,omitempty"`
	After          *time.Time `json:"after,omitempty"`
	Limit          int        `json:"limit,omitempty"`
	IncludePrivate bool       `json:"includePrivate,omitempty"`
}

type MessageUpdate struct {
	Content string      `json:"content,omitempty"`
	Embeds  []ChatEmbed `json:"embeds,omitempty"`
}
