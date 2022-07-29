package guildrone

import (
	"time"

	"github.com/go-playground/validator/v10"
)

const (
	MessageTypeDefault MessageType = "default"
	MessageTypeSystem  MessageType = "system"
)

// ChatMessage stores the data of a chat message
type ChatMessage struct {
	ID                 string      `json:"id"`
	Type               MessageType `json:"type"`
	ServerID           string      `json:"serverId"`
	ChannelID          string      `json:"channelId"`
	Content            string      `json:"content"`
	Embeds             []ChatEmbed `json:"embeds" validate:"omitempty,min=1,max=10,dive"`
	ReplyMessageIds    []string    `json:"replyMessageIds" validate:"omitempty,min=1,max=5"`
	IsPrivate          bool        `json:"isPrivate"`
	IsSilent           bool        `json:"isSilent"`
	Mentions           *Mentions   `json:"mentions,omitempty" validate:"omitempty,dive"`
	CreatedAt          time.Time   `json:"createdAt"`
	CreatedBy          string      `json:"createdBy"`
	CreatedByWebhookId string      `json:"createdByWebhookId"`
	UpdatedAt          *time.Time  `json:"updatedAt,omitempty"`
}

// MessageType is the type of message
type MessageType string

// ChatEmbed stores the data of an embed
type ChatEmbed struct {
	Title       string              `json:"title,omitempty" validate:"max=256"`
	Description string              `json:"description,omitempty" validate:"max=2048"`
	URL         string              `json:"url,omitempty" validate:"url,max=1024"`
	Color       int                 `json:"color" validate:"min=0,max=16777215"`
	Footer      *ChatEmbedFooter    `json:"footer,omitempty" validate:"omitempty,dive"`
	Timestamp   *time.Time          `json:"timestamp,omitempty" validate:"omitempty,dive"`
	Thumbnail   *ChatEmbedThumbnail `json:"thumbnail,omitempty" validate:"omitempty,dive"`
	Image       *ChatEmbedImage     `json:"image,omitempty" validate:"omitempty,dive"`
	Author      *ChatEmbedAuthor    `json:"author,omitempty" validate:"omitempty,dive"`
	Fields      []ChatEmbedField    `json:"fields,omitempty" validate:"omitempty,max=25,dive"`
}

// Validate validates the ChatEmbed
// Returns nil if no errors were found
func (e *ChatEmbed) Validate() error {
	validate := validator.New()
	return validate.Struct(e)
}

type ChatEmbedFooter struct {
	Text    string `json:"text,omitempty" validate:"max=2048"`
	IconUrl string `json:"icon_url,omitempty" validate:"url,max=1024"`
}

type ChatEmbedThumbnail struct {
	URL string `json:"url,omitempty" validate:"url,max=1024"`
}

type ChatEmbedImage struct {
	URL string `json:"url,omitempty" validate:"url,max=1024"`
}

type ChatEmbedAuthor struct {
	Name    string `json:"name,omitempty" validate:"max=256"`
	URL     string `json:"url,omitempty" validate:"url,max=1024"`
	IconUrl string `json:"icon_url,omitempty" validate:"url,max=1024"`
}

type ChatEmbedField struct {
	Name   string `json:"name" validate:"max=256"`
	Value  string `json:"value" validate:"max=1024"`
	Inline bool   `json:"inline,omitempty"`
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

// MessageCreate is a request body for creating a message
type MessageCreate struct {
	IsPrivate       bool        `json:"isPrivate,omitempty"`
	IsSilent        bool        `json:"isSilent,omitempty"`
	ReplyMessageIds []string    `json:"replyMessageIds,omitempty" validate:"omitempty,min=1,max=5"`
	Content         string      `json:"content,omitempty" validate:"max=4000"`
	Embeds          []ChatEmbed `json:"embeds,omitempty" validate:"omitempty,dive"`
}

// Validate validates the MessageCreate request body
// Returns nil if no errors were found
func (m *MessageCreate) Validate() error {
	validate := validator.New()
	return validate.Struct(m)
}

// MessagesRequest is a request body for getting messages
type MessagesRequest struct {
	Before         *time.Time `json:"before,omitempty"`
	After          *time.Time `json:"after,omitempty"`
	Limit          int        `json:"limit,omitempty"`
	IncludePrivate bool       `json:"includePrivate,omitempty"`
}

// MessageUpdate is a request body for updating a message
type MessageUpdate struct {
	Content string      `json:"content,omitempty" validate:"max=4000"`
	Embeds  []ChatEmbed `json:"embeds,omitempty" validate:"omitempty,dive"`
}

// Validate validates the MessageUpdate request body
// Returns nil if no errors were found
func (m *MessageUpdate) Validate() error {
	validate := validator.New()
	return validate.Struct(m)
}
