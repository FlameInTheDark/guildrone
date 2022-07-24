package guildrone

import (
	"encoding/json"
	"time"
)

// This file contains all the possible structs that can be
// handled by AddHandler/EventHandler.
// DO NOT ADD ANYTHING BUT EVENT HANDLER STRUCTS TO THIS FILE.
//go:generate go run tools/cmd/eventhandlers/main.go

// Connect is the data for a Connect event.
// This is a synthetic event and is not dispatched by Guilded.
type Connect struct{}

// Disconnect is the data for a Disconnect event.
// This is a synthetic event and is not dispatched by Guilded.
type Disconnect struct{}

// Resume is the data for a Resume event.
// This is a synthetic event and is not dispatched by Guilded.
type Resume struct{}

// RateLimit is the data for a RateLimit event.
// This is a synthetic event and is not dispatched by Guilded.
type RateLimit struct {
	RetryAfter time.Duration
	URL        string
}

// Event provides a basic initial struct for all websocket events.
type Event struct {
	Operation int             `json:"op"`
	MessageID string          `json:"s"`
	Type      string          `json:"t"`
	RawData   json.RawMessage `json:"d"`
	// Struct contains one of the other types in this file.
	Struct interface{} `json:"-"`
}

// Is the data for the ChatMessageCreated event
type ChatMessageCreated struct {
	ServerID string      `json:"serverId"`
	Message  ChatMessage `json:"message"`
}

type ChatMessageUpdated struct {
	ServerID string      `json:"serverId"`
	Message  ChatMessage `json:"message"`
}

type ChatMessageDeleted struct {
	ServerID string      `json:"serverId"`
	Message  ChatMessage `json:"message"`
}

type TeamMemberJoined struct {
	ServerID string       `json:"serverId"`
	Member   ServerMember `json:"member"`
}

type TeamMemberRemoved struct {
	ServerID string `json:"serverId"`
	UserID   string `json:"userId"`
	IsKick   bool   `json:"isKick"`
	IsBan    bool   `json:"isBan"`
}

type TeamMemberBanned struct {
	ServerID        string          `json:"serverId"`
	ServerMemberBan ServerMemberBan `json:"serverMemberBan"`
}

type TeamMemberUnbanned struct {
	ServerID        string          `json:"serverId"`
	ServerMemberBan ServerMemberBan `json:"serverMemberBan"`
}

type TeamMemberUpdated struct {
	ServerID string   `json:"serverId"`
	UserInfo UserInfo `json:"userInfo"`
}

type TeamRolesUpdated struct {
	ServerID      string       `json:"serverId"`
	MemberRoleIds []MemberRole `json:"memberRoleIds"`
}

type TeamChannelCreated struct {
	ServerID string        `json:"serverId"`
	Channel  ServerChannel `json:"channel"`
}

type TeamChannelUpdated struct {
	ServerID string        `json:"serverId"`
	Channel  ServerChannel `json:"channel"`
}

type TeamWebhookCreated struct {
	ServerID string  `json:"serverId"`
	Webhook  Webhook `json:"webhook"`
}

type TeamWebhookUpdated struct {
	ServerID string  `json:"serverId"`
	Webhook  Webhook `json:"webhook"`
}

type DocCreated struct {
	ServerID string `json:"serverId"`
	Doc      Doc    `json:"doc"`
}

type DocUpdated struct {
	ServerID string `json:"serverId"`
	Doc      Doc    `json:"doc"`
}

type DocDeleted struct {
	ServerID string `json:"serverId"`
	Doc      Doc    `json:"doc"`
}

type CalendarEventCreated struct {
	ServerID      string        `json:"serverId"`
	CalendarEvent CalendarEvent `json:"calendarEvent"`
}

type CalendarEventUpdated struct {
	ServerID      string        `json:"serverId"`
	CalendarEvent CalendarEvent `json:"calendarEvent"`
}

type CalendarEventDeleted struct {
	ServerID      string        `json:"serverId"`
	CalendarEvent CalendarEvent `json:"calendarEvent"`
}

type ListItemCreated struct {
	ServerID string   `json:"serverId"`
	ListItem ListItem `json:"listItem"`
}

type ListItemUpdated struct {
	ServerID string   `json:"serverId"`
	ListItem ListItem `json:"listItem"`
}

type ListItemDeleted struct {
	ServerID string   `json:"serverId"`
	ListItem ListItem `json:"listItem"`
}

type ListItemCompleted struct {
	ServerID string   `json:"serverId"`
	ListItem ListItem `json:"listItem"`
}

type ChannelMessageReactionCreated struct {
	ServerID string   `json:"serverId"`
	Reaction Reaction `json:"reaction"`
}

type ChannelMessageReactionDeleted struct {
	ServerID string   `json:"serverId"`
	Reaction Reaction `json:"reaction"`
}

type Ready struct {
	LastMessageID       string  `json:"lastMessageId"`
	HeartbeatIntervalMS int     `json:"heartbeatIntervalMs"`
	User                BotUser `json:"user"`
}

type ForumTopicCreated struct {
	ServerID   string     `json:"serverId"`
	ForumTopic ForumTopic `json:"forumTopic"`
}

type ForumTopicUpdated struct {
	ServerID   string     `json:"serverId"`
	ForumTopic ForumTopic `json:"forumTopic"`
}

type ForumTopicDeleted struct {
	ServerID   string     `json:"serverId"`
	ForumTopic ForumTopic `json:"forumTopic"`
}

type CalendarEventRsvpUpdated struct {
	ServerID          string            `json:"serverId"`
	CalendarEventRsvp CalendarEventRsvp `json:"calendarEventRsvp"`
}

type CalendarEventRsvpManyUpdated struct {
	ServerID           string              `json:"serverId"`
	CalendarEventRsvps []CalendarEventRsvp `json:"calendarEventRsvps"`
}

type CalendarEventRsvpDeleted struct {
	ServerID          string            `json:"serverId"`
	CalendarEventRsvp CalendarEventRsvp `json:"calendarEventRsvp"`
}
