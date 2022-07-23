package guildrone

import (
	"net/http"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/websocket"
)

const (
	ServerChannelTypeAnnouncements ServerChannelType = "announcements"
	ServerChannelTypeChat          ServerChannelType = "chat"
	ServerChannelTypeCalendar      ServerChannelType = "calendar"
	ServerChannelTypeForums        ServerChannelType = "forums"
	ServerChannelTypeMedia         ServerChannelType = "media"
	ServerChannelTypeDocs          ServerChannelType = "docs"
	ServerChannelTypeVoice         ServerChannelType = "voice"
	ServerChannelTypeList          ServerChannelType = "list"
	ServerChannelTypeScheduling    ServerChannelType = "scheduling"
	ServerChannelTypeStream        ServerChannelType = "stream"
)

const (
	UserTypeUser UserType = "user"
	UserTypeBot  UserType = "bot"
)

type Session struct {
	sync.RWMutex

	// Authentication token for this session
	Token string

	// Logging
	Debug    bool
	LogLevel int

	// REST API Client
	Client    *http.Client
	UserAgent string

	// Max number of REST API retries
	MaxRestRetries int

	// Should the session reconnect the websocket on errors.
	ShouldReconnectOnError bool

	// ID of the last websocket event message
	eventMu     sync.RWMutex
	LastEventID string

	// Should replay missed events on websocket reconnect
	ShouldReplayEventsOnReconnect bool

	// Should the session retry requests when rate limited.
	ShouldRetryOnRateLimit bool

	// Whether or not to call event handlers synchronously.
	// e.g false = launch event handlers in their own goroutines.
	SyncEvents bool

	// Whether the Data Websocket is ready
	DataReady bool // NOTE: Maye be deprecated soon

	// Event handlers
	handlersMu   sync.RWMutex
	handlers     map[string][]*eventHandlerInstance
	onceHandlers map[string][]*eventHandlerInstance

	// The websocket connection.
	wsConn *websocket.Conn

	// When nil, the session is not listening.
	listening chan interface{}

	// Stores the last HeartbeatAck that was received (in UTC)
	LastHeartbeatAck time.Time

	// Stores the last Heartbeat sent (in UTC)
	LastHeartbeatSent time.Time

	// used to make sure gateway websocket writes do not happen concurrently
	wsMutex sync.Mutex
}

type ServerMember struct {
	User     User   `json:"user"`
	RoleIds  []int  `json:"roleIds"`
	Nickname string `json:"nickname"`
	JoinedAt string `json:"joinedAt"`
	IsOwner  bool   `json:"isOwner"`
}

type UserType string

type UserSummary struct {
	ID     string `json:"id"`
	Type   string `json:"type"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

type UserInfo struct {
	ID       string `json:"id"`
	Nickname string `json:"nickname"`
}

type ServerMemberBan struct {
	UserSummary UserSummary `json:"user"`
	Reason      string      `json:"reason"`
	CreatedBy   string      `json:"createdBy"`
	CreatedAt   time.Time   `json:"createdAt"`
}

type MemberRole struct {
	UserID  string `json:"userId"`
	RoleIDs []int  `json:"roleId"`
}

// ServerChannelType is a sting that represents the type of a server channel.
// ("announcements", "chat", "calendar", "forums", "media", "docs", "voice", "list", "scheduling", or "stream")
type ServerChannelType string

type ServerChannel struct {
	ID         string            `json:"id"`
	Type       ServerChannelType `json:"type"`
	Name       string            `json:"name"`
	Topic      string            `json:"topic"`
	CreatedAt  time.Time         `json:"createdAt"`
	CreatedBy  string            `json:"createdBy"`
	UpdatedAt  *time.Time        `json:"updatedAt"`
	ServerId   string            `json:"serverId"`
	ParentId   string            `json:"parentId"`
	CategoryId int               `json:"categoryId"`
	GroupId    string            `json:"groupId"`
	IsPublic   bool              `json:"isPublic"`
	ArchivedBy string            `json:"archivedBy"`
	ArchivedAt string            `json:"archivedAt"`
}

type Webhook struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	ServerId  string     `json:"serverId"`
	ChannelId string     `json:"channelId"`
	CreatedAt time.Time  `json:"createdAt"`
	CreatedBy string     `json:"createdBy"`
	DeletedAt *time.Time `json:"deletedAt"`
	Token     string     `json:"token"`
}

type Doc struct {
	ID        int        `json:"id"`
	ServerId  string     `json:"serverId"`
	ChannelId string     `json:"channelId"`
	Title     string     `json:"title"`
	Content   string     `json:"content"`
	Mentions  *Mentions  `json:"mentions"`
	CreatedAt time.Time  `json:"createdAt"`
	CreatedBy string     `json:"createdBy"`
	UpdatedAt *time.Time `json:"updatedAt"`
	UpdatedBy string     `json:"updatedBy"`
}

type CalendarEvent struct {
	ID          string    `json:"id"`
	ServerId    string    `json:"serverId"`
	ChannelId   string    `json:"channelId"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Location    string    `json:"location"`
	URL         string    `json:"url"`
	Color       int       `json:"color"`
	StartsAt    time.Time `json:"startsAt"`
	// Duration in minutes
	Duration     int          `json:"duration"`
	IsPrivate    bool         `json:"isPrivate"`
	Mentions     *Mentions    `json:"mentions"`
	CreatedAt    time.Time    `json:"createdAt"`
	CreatedBy    string       `json:"createdBy"`
	Cancellation Cancellation `json:"cancellation"`
}

type Cancellation struct {
	Description string `json:"description"`
	CreatedBy   string `json:"createdBy"`
}

// ListItem is a struct that represents a list item.
type ListItem struct {
	ID                 string        `json:"id"`
	ServerID           string        `json:"serverId"`
	ChannelID          string        `json:"channelId"`
	Message            string        `json:"message"`
	Mentions           *Mentions     `json:"mentions"`
	CreatedAt          time.Time     `json:"createdAt"`
	CreatedBy          string        `json:"createdBy"`
	CreatedByWebhookId string        `json:"createdByWebhookId"`
	UpdatedAt          *time.Time    `json:"updatedAt"`
	UpdatedBy          string        `json:"updatedBy"`
	ParentListItemId   string        `json:"parentListItemId"`
	CompletedAt        string        `json:"completedAt"`
	CompletedBy        string        `json:"completedBy"`
	Note               *ListItemNote `json:"note"`
}

// ListItemNote is a struct that represents a list item note.
type ListItemNote struct {
	CreatedAt time.Time  `json:"createdAt"`
	CreatedBy string     `json:"createdBy"`
	UpdatedAt *time.Time `json:"updatedAt"`
	UpdatedBy string     `json:"updatedBy"`
	Mentions  *Mentions  `json:"mentions"`
	Content   string     `json:"content"`
}

// Reaction is a struct that represents a reaction.
type Reaction struct {
	ChannelID string `json:"channelId"`
	MessageID string `json:"messageId"`
	CreatedBy string `json:"createdBy"`
	Emote     Emote  `json:"emote"`
}

// Emote is a struct that represents a reaction emote.
type Emote struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

// BotUser is a bot data structure.
type BotUser struct {
	ID        string    `json:"id"`
	BotID     string    `json:"botId"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	CreatedBy string    `json:"createdBy"`
}

// An APIErrorMessage is an api error message returned from Guilded
type APIErrorMessage struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

const (
	ServerTypeTeam         ServerType = "team"
	ServerTypeOrganization ServerType = "organization"
	ServerTypeCommunity    ServerType = "community"
	ServerTypeClan         ServerType = "clan"
	ServerTypeGuild        ServerType = "guild"
	ServerTypeFriends      ServerType = "friends"
	ServerTypeStreaming    ServerType = "streaming"
	ServerTypeOther        ServerType = "other"
)

type ServerType string

// Server represents a server in Guilded
type Server struct {
	ID               string      `json:"id"`
	OwnerID          string      `json:"ownerId"`
	Type             *ServerType `json:"type,omitempty"`
	Name             string      `json:"name"`
	URL              string      `json:"url,omitempty"`
	About            string      `json:"about,omitempty"`
	Avatar           string      `json:"avatar,omitempty"`
	Banner           string      `json:"banner,omitempty"`
	Timezone         string      `json:"timezone,omitempty"`
	IsVerified       bool        `json:"isVerified,omitempty"`
	DefaultChannelID string      `json:"defaultChannelId,omitempty"`
	CreatedAt        time.Time   `json:"createdAt"`
}

// ServerChannelCreate is the request body for creating a channel
type ServerChannelCreate struct {
	Name       string            `json:"name" validate:"required,min=1,max=100"`
	Topic      string            `json:"topic,omitempty" validate:"omitempty,min=1,max=512"`
	IsPublic   bool              `json:"isPublic,omitempty"`
	Type       ServerChannelType `json:"type" validate:"required"`
	ServerID   string            `json:"serverId,omitempty"`
	GroupID    string            `json:"groupId,omitempty"`
	CategoryID string            `json:"categoryId,omitempty"`
}

// Validate validates the channel create request
// Returns nil if valid, otherwise returns an error
func (c *ServerChannelCreate) Validate() error {
	validate := validator.New()
	return validate.Struct(c)
}

// ServerChannelUpdate is the request body for updating a channel
type ServerChannelUpdate struct {
	Name     string `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Topic    string `json:"topic,omitempty" validate:"omitempty,min=1,max=512"`
	IsPublic bool   `json:"isPublic,omitempty"`
}

// Validate validates the channel update request
// Returns nil if valid, otherwise returns an error
func (c *ServerChannelUpdate) Validate() error {
	validate := validator.New()
	return validate.Struct(c)
}

// ServerMemberNicknameUpdate is the request body for updating a server member nickname
type ServerMemberNicknameUpdate struct {
	Nickname string `json:"nickname"`
}

// ServerMemberBanCreate is the request body for banning a user from a server
type ServerMemberBanCreate struct {
	Reason string `json:"reason"`
}

// ForumTopic is the forum topic model
type ForumTopic struct {
	ID                 int        `json:"id"`
	ServerID           string     `json:"serverId"`
	ChannelID          string     `json:"channelId"`
	Title              string     `json:"title,omitempty"`
	Content            string     `json:"content,omitempty"`
	CreatedAt          time.Time  `json:"createdAt"`
	CreatedBy          string     `json:"createdBy"`
	CreatedByWebhookId string     `json:"createdByWebhookId,omitempty"`
	UpdatedAt          *time.Time `json:"updatedAt,omitempty"`
}

// ForumTopicSummary is the forum topic summary model
type ForumTopicSummary struct {
	ID               int        `json:"id"`
	ServerID         string     `json:"serverId"`
	ChannelID        string     `json:"channelId"`
	Title            string     `json:"title"`
	CreatedAt        time.Time  `json:"createdAt"`
	CreatedBy        string     `json:"createdBy"`
	CreatedByWebhook string     `json:"createdByWebhook,omitempty"`
	UpdatedAt        *time.Time `json:"updatedAt,omitempty"`
	BumpedAt         *time.Time `json:"bumpedAt,omitempty"`
}

// ChannelForumTopicCreate is the request body for creating a forum topic
type ChannelForumTopicCreate struct {
	Title   string `json:"title" validate:"min=1"`
	Content string `json:"content"`
}

// Validate validates the channel forum topic create request
// Returns nil if valid, otherwise returns an error
func (c *ChannelForumTopicCreate) Validate() error {
	validate := validator.New()
	return validate.Struct(c)
}

// ChannelForumTopicUpdate is the request body for updating a forum topic
type ChannelForumTopicUpdate struct {
	Title   string `json:"title,omitempty" validate:"min=1,max=500"`
	Content string `json:"content,omitempty"`
}

// Validate validates the channel forum topic update request
func (c *ChannelForumTopicUpdate) Validate() error {
	validate := validator.New()
	return validate.Struct(c)
}

// ChannelListItem is the request body for creating or updating a channel list item
type ChannelListItem struct {
	Message string               `json:"message"`
	Note    *ChannelListItemNote `json:"note,omitempty"`
}

type ChannelListItemNote struct {
	Content string `json:"content"`
}

type ChannelDoc struct {
	Title   string `json:"title" validate:"min=1"`
	Content string `json:"content"`
}

// Validate validates the channel doc create/update request
// Returns nil if valid, otherwise returns an error
func (c *ChannelDoc) Validate() error {
	validate := validator.New()
	return validate.Struct(c)
}

// ChannelEvent is the request body for creating/updating a channel event
type ChannelEvent struct {
	Name        string     `json:"name" validate:"min=1,max=60"`
	Description string     `json:"description,omitempty" validate:"omitempty,min=1,max=8000"`
	Location    string     `json:"location,omitempty" validate:"omitempty,min=1,max=8000"`
	StartsAt    *time.Time `json:"startsAt,omitempty"`
	URL         string     `json:"url,omitempty"`
	Color       int        `json:"color,omitempty" validate:"omitempty,min=0,max=16777215"`
	Duration    int        `json:"duration,omitempty" validate:"omitempty,min=1"`
	IsPrivate   bool       `json:"isPrivate,omitempty"`
}

// Validate validates the channel event create/update request
// Returns nil if valid, otherwise returns an error
func (e *ChannelEvent) Validate() error {
	validate := validator.New()
	return validate.Struct(e)
}

// ChannelEventsRequest is the request body for listing channel events
type ChannelEventsRequest struct {
	Before *time.Time `json:"before,omitempty"`
	After  *time.Time `json:"after,omitempty"`
	Limit  int        `json:"limit,omitempty" validate:"omitempty,min=1,max=500"`
}

// Validate validates the channel events request
// Returns nil if valid, otherwise returns an error
func (e *ChannelEventsRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(e)
}

// ServerXPUpdate is the request body for updating a server member xp
type ServerXPUpdate struct {
	Amount int `json:"amount" validate:"min=-1000,max=1000"`
}

// Validate validates the server member xp update request
// Returns nil if valid, otherwise returns an error
func (xp *ServerXPUpdate) Validate() error {
	validate := validator.New()
	return validate.Struct(xp)
}

// ServerSocialLink is the request body for retrieving server member social link
type ServerSocialLink struct {
	Handle    string `json:"handle,omitempty"`
	ServiceID string `json:"serviceId,omitempty"`
	Type      string `json:"type"`
}

// WebhookCreate is the request body for creating a webhook
type WebhookCreate struct {
	Name      string `json:"name" validate:"min=1,max=128"`
	ChannelID string `json:"channelId"`
}

// Validate validates the webhook create request
// Returns nil if valid, otherwise returns an error
func (w *WebhookCreate) Validate() error {
	validate := validator.New()
	return validate.Struct(w)
}

// WebhookUpdate is the request body for updating a webhook
type WebhookUpdate struct {
	Name      string `json:"name" validate:"min=1,max=128"`
	ChannelID string `json:"channelId"`
}

// Validate validates the webhook update request
// Returns nil if valid, otherwise returns an error
func (w *WebhookUpdate) Validate() error {
	validate := validator.New()
	return validate.Struct(w)
}
