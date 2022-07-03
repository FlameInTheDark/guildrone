package guildrone

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
	"time"
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
	CreatedAt   Timestamp   `json:"createdAt"`
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
	CreatedAt  Timestamp         `json:"createdAt"`
	CreatedBy  string            `json:"createdBy"`
	UpdatedAt  *Timestamp        `json:"updatedAt"`
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
	CreatedAt Timestamp  `json:"createdAt"`
	CreatedBy string     `json:"createdBy"`
	DeletedAt *Timestamp `json:"deletedAt"`
	Token     string     `json:"token"`
}

type Doc struct {
	ID        int        `json:"id"`
	ServerId  string     `json:"serverId"`
	ChannelId string     `json:"channelId"`
	Title     string     `json:"title"`
	Content   string     `json:"content"`
	Mentions  *Mentions  `json:"mentions"`
	CreatedAt Timestamp  `json:"createdAt"`
	CreatedBy string     `json:"createdBy"`
	UpdatedAt *Timestamp `json:"updatedAt"`
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
	StartsAt    Timestamp `json:"startsAt"`
	// Duration in minutes
	Duration     int          `json:"duration"`
	IsPrivate    bool         `json:"isPrivate"`
	Mentions     *Mentions    `json:"mentions"`
	CreatedAt    Timestamp    `json:"createdAt"`
	CreatedBy    string       `json:"createdBy"`
	Cancellation Cancellation `json:"cancellation"`
}

type Cancellation struct {
	Description string `json:"description"`
	CreatedBy   string `json:"createdBy"`
}

type ListItem struct {
	ID                 string        `json:"id"`
	ServerID           string        `json:"serverId"`
	ChannelID          string        `json:"channelId"`
	Message            string        `json:"message"`
	Mentions           *Mentions     `json:"mentions"`
	CreatedAt          Timestamp     `json:"createdAt"`
	CreatedBy          string        `json:"createdBy"`
	CreatedByWebhookId string        `json:"createdByWebhookId"`
	UpdatedAt          *Timestamp    `json:"updatedAt"`
	UpdatedBy          string        `json:"updatedBy"`
	ParentListItemId   string        `json:"parentListItemId"`
	CompletedAt        string        `json:"completedAt"`
	CompletedBy        string        `json:"completedBy"`
	Note               *ListItemNote `json:"note"`
}

type ListItemNote struct {
	CreatedAt Timestamp  `json:"createdAt"`
	CreatedBy string     `json:"createdBy"`
	UpdatedAt *Timestamp `json:"updatedAt"`
	UpdatedBy string     `json:"updatedBy"`
	Mentions  *Mentions  `json:"mentions"`
	Content   string     `json:"content"`
}

type Reaction struct {
	ChannelID string `json:"channelId"`
	MessageID string `json:"messageId"`
	CreatedBy string `json:"createdBy"`
	Emote     Emote  `json:"emote"`
}

type Emote struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

type Timestamp struct {
	RawTime string    `json:"rawTime"`
	Time    time.Time `json:"time"`
}

func (t Timestamp) String() string {
	if t.RawTime != "" {
		return t.RawTime
	}
	return t.Time.Format(time.RFC3339)
}

func (t Timestamp) MarshalJSON() ([]byte, error) {
	if t.RawTime != "" {
		return []byte(t.RawTime), nil
	}
	return []byte(t.Time.Format(time.RFC3339)), nil
}

func (t *Timestamp) UnmarshalJSON(b []byte) error {
	var s string
	var err error
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	t.RawTime = s
	t.Time, err = time.Parse(time.RFC3339, s)
	return err
}

type BotUser struct {
	ID        string    `json:"id"`
	BotID     string    `json:"botId"`
	Name      string    `json:"name"`
	CreatedAt Timestamp `json:"createdAt"`
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
	DefaultChannleID string      `json:"defaultChannelId,omitempty"`
	CreatedAt        Timestamp   `json:"createdAt"`
}

type ChannelCreate struct {
	Name       string            `json:"name"`
	Topic      string            `json:"topic,omitempty"`
	IsPublic   bool              `json:"isPublic,omitempty"`
	Type       ServerChannelType `json:"type"`
	ServerID   string            `json:"serverId,omitempty"`
	GroupID    string            `json:"groupId,omitempty"`
	CategoryID string            `json:"categoryId,omitempty"`
}

type ChannelUpdate struct {
	Name     string `json:"name,omitempty"`
	Topic    string `json:"topic,omitempty"`
	IsPublic bool   `json:"isPublic,omitempty"`
}

type ServerMemberNicknameUpdate struct {
	Nickname string `json:"nickname"`
}

type ServerMemberBanCreate struct {
	Reason string `json:"reason"`
}

type ForumTopic struct {
	ID                 int        `json:"id"`
	ServerID           string     `json:"serverId"`
	ChannelID          string     `json:"channelId"`
	Title              string     `json:"title,omitempty"`
	Content            string     `json:"content,omitempty"`
	CreatedAt          Timestamp  `json:"createdAt"`
	CreatedBy          string     `json:"createdBy"`
	CreatedByWebhookId string     `json:"createdByWebhookId,omitempty"`
	UpdatedAt          *Timestamp `json:"updatedAt,omitempty"`
}

type ChannelForumTopicCreate struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type ChannelListItem struct {
	Message string               `json:"message"`
	Note    *ChannelListItemNote `json:"note,omitempty"`
}

type ChannelListItemNote struct {
	Content string `json:"content"`
}

type ChannelDoc struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type ChannelEvent struct {
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	Location    string     `json:"location,omitempty"`
	StartsAt    *Timestamp `json:"startsAt,omitempty"`
	URL         string     `json:"url,omitempty"`
	Color       int        `json:"color,omitempty"`
	Duration    int        `json:"duration,omitempty"`
	IsPrivate   bool       `json:"isPrivate,omitempty"`
}

type ChannelEventsRequest struct {
	Before *Timestamp `json:"before,omitempty"`
	After  *Timestamp `json:"after,omitempty"`
	Limit  int        `json:"limit,omitempty"`
}

type ServerXPUpdate struct {
	Amount int `json:"amount"`
}

type ServerSocialLink struct {
	Handle    string `json:"handle,omitempty"`
	ServiceID string `json:"serviceId,omitempty"`
	Type      string `json:"type"`
}

type WebhookCreate struct {
	Name      string `json:"name"`
	ChannelID string `json:"channelId"`
}

type WebhookUpdate struct {
	Name      string `json:"name"`
	ChannelID string `json:"channelId"`
}
