package guildrone

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// All error constants
var (
	ErrJSONUnmarshal           = errors.New("json unmarshal")
	ErrStatusOffline           = errors.New("You can't set your Status to offline")
	ErrVerificationLevelBounds = errors.New("VerificationLevel out of bounds, should be between 0 and 3")
	ErrPruneDaysBounds         = errors.New("the number of days should be more than or equal to 1")
	ErrGuildNoIcon             = errors.New("guild does not have an icon set")
	ErrGuildNoSplash           = errors.New("guild does not have a splash set")
	ErrUnauthorized            = errors.New("HTTP request was unauthorized. This could be because the provided token was not a bot token")
)

var (
	// Marshal defines function used to encode JSON payloads
	Marshal func(v interface{}) ([]byte, error) = json.Marshal
	// Unmarshal defines function used to decode JSON payloads
	Unmarshal func(src []byte, v interface{}) error = json.Unmarshal
)

// RESTError stores error information about a request with a bad response code.
// Message is not always present, there are cases where api calls can fail
// without returning a json message.
type RESTError struct {
	Request      *http.Request
	Response     *http.Response
	ResponseBody []byte

	Message *APIErrorMessage // Message may be nil.
}

// newRestError returns a new REST API error.
func newRestError(req *http.Request, resp *http.Response, body []byte) *RESTError {
	restErr := &RESTError{
		Request:      req,
		Response:     resp,
		ResponseBody: body,
	}

	// Attempt to decode the error and assume no message was provided if it fails
	var msg *APIErrorMessage
	err := Unmarshal(body, &msg)
	if err == nil {
		restErr.Message = msg
	}

	return restErr
}

// Error returns a Rest API Error with its status code and body.
func (r RESTError) Error() string {
	return "HTTP " + r.Response.Status + ", " + string(r.ResponseBody)
}

// RateLimitError is returned when a request exceeds a rate limit
// and ShouldRetryOnRateLimit is false. The request may be manually
// retried after waiting the duration specified by RetryAfter.
type RateLimitError struct {
	*RateLimit
}

// Error returns a rate limit error with rate limited endpoint and retry time.
func (e RateLimitError) Error() string {
	return "Rate limit exceeded on " + e.URL + ", retry after " + e.RetryAfter.String()
}

// Request is the same as RequestWithBucketID but the bucket id is the same as the urlStr
func (s *Session) Request(method, urlStr string, data interface{}) (response []byte, err error) {
	var body []byte
	if data != nil {
		body, err = Marshal(data)
		if err != nil {
			return
		}
	}

	return s.request(method, urlStr, "application/json", body, 0)
}

// request makes a (GET/POST/...) Requests to Guilded REST API.
// Sequence is the sequence number, if it fails with a 502 it will
// retry with sequence+1 until it either succeeds or sequence >= session.MaxRestRetries
func (s *Session) request(method, urlStr, contentType string, b []byte, sequence int) (response []byte, err error) {
	return s.RequestCall(method, urlStr, contentType, b, sequence)
}

// RequestCall makes a request using a bucket that's already been locked
func (s *Session) RequestCall(method, urlStr, contentType string, b []byte, sequence int) (response []byte, err error) {
	if s.Debug {
		log.Printf("API REQUEST %8s :: %s\n", method, urlStr)
		log.Printf("API REQUEST  PAYLOAD :: [%s]\n", string(b))
	}

	req, err := http.NewRequest(method, urlStr, bytes.NewBuffer(b))
	if err != nil {
		//bucket.Release(nil)
		return
	}

	// Not used on initial login..
	// TODO: Verify if a login, otherwise complain about no-token
	if s.Token != "" {
		req.Header.Set("Authorization", "Bearer "+s.Token)
	}

	// Set the content type if the request has a body
	if b != nil {
		req.Header.Set("Content-Type", contentType)
	}

	// TODO: Make a configurable static variable.
	req.Header.Set("User-Agent", s.UserAgent)

	if s.Debug {
		for k, v := range req.Header {
			log.Printf("API REQUEST   HEADER :: [%s] = %+v\n", k, v)
		}
	}

	resp, err := s.Client.Do(req)
	if err != nil {
		//bucket.Release(nil)
		return
	}
	defer func() {
		err2 := resp.Body.Close()
		if s.Debug && err2 != nil {
			log.Println("error closing resp body")
		}
	}()

	//err = bucket.Release(resp.Header)
	//if err != nil {
	//	return
	//}

	response, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	if s.Debug {

		log.Printf("API RESPONSE  STATUS :: %s\n", resp.Status)
		for k, v := range resp.Header {
			log.Printf("API RESPONSE  HEADER :: [%s] = %+v\n", k, v)
		}
		log.Printf("API RESPONSE    BODY :: [%s]\n\n\n", response)
	}

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusCreated:
	case http.StatusNoContent:
	case http.StatusBadGateway:
		// Retry sending request if possible
		if sequence < s.MaxRestRetries {

			s.log(LogInformational, "%s Failed (%s), Retrying...", urlStr, resp.Status)
			response, err = s.RequestCall(method, urlStr, contentType, b, sequence+1)
		} else {
			err = fmt.Errorf("Exceeded Max retries HTTP %s, %s", resp.Status, response)
		}
	case 429: // TOO MANY REQUESTS - Rate limiting
		after, err := strconv.ParseInt(resp.Header.Get("Retry-After"), 10, 64)
		if err != nil {
			return nil, err
		}

		if s.ShouldRetryOnRateLimit {

			s.log(LogInformational, "Rate Limiting %s, retry in %d", urlStr, after)
			s.handleEvent(rateLimitEventType, &RateLimit{RetryAfter: time.Duration(after), URL: urlStr})

			time.Sleep(time.Duration(after))

			response, err = s.RequestCall(method, urlStr, contentType, b, sequence)
		} else {
			err = &RateLimitError{&RateLimit{RetryAfter: time.Duration(after), URL: urlStr}}
		}
	case http.StatusUnauthorized:
		if strings.Index(s.Token, "Bot ") != 0 {
			s.log(LogInformational, ErrUnauthorized.Error())
			err = ErrUnauthorized
		}
		fallthrough
	default: // Error condition
		err = newRestError(req, resp, response)
	}

	return
}

func unmarshal(data []byte, v interface{}) error {
	err := Unmarshal(data, v)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrJSONUnmarshal, err)
	}

	return nil
}

// ------------------------------------------------------------------------------------------------
// Functions specific to Guilded Messages
// ------------------------------------------------------------------------------------------------

// ChannelMessageCreateComplex sends a message to the given channel.
// channelID : The ID of a Channel.
// data      : The message struct to send.
func (s *Session) ChannelMessageCreateComplex(channelID string, data *MessageCreate) (st *ChatMessage, err error) {
	body, err := s.Request("POST", EndpointChannelMessages(channelID), data)
	if err != nil {
		return
	}

	err = unmarshal(body, &st)
	return
}

// ChannelMessageCreate sends a message to the given channel.
// channelID : The ID of a Channel.
// content   : The message to send.
func (s *Session) ChannelMessageCreate(channelID string, content string) (*ChatMessage, error) {
	return s.ChannelMessageCreateComplex(channelID, &MessageCreate{
		Content: content,
	})
}

// ChannelMessage returns a message from a channel.
// channelID : The ID of a Channel.
// messageID : The ID of a Message.
func (s *Session) ChannelMessage(channelID string, messageID string) (*ChatMessage, error) {
	body, err := s.Request("GET", EndpointChannelMessage(channelID, messageID), nil)
	if err != nil {
		return nil, err
	}

	var st *ChatMessage
	err = unmarshal(body, &st)
	return st, err
}

// ChannelMessages returns an array of messages from a channel.
// channelID      : The ID of a Channel.
// limit          : The number of messages to return.
// beforeTime     : The time before which messages are to be returned.
// afterTime      : The time after which messages are to be returned.
// includePrivate : Whether to include private messages.
func (s *Session) ChannelMessages(channelID string, limit int, beforeTime, afterTime *time.Time, includePrivate bool) ([]*ChatMessage, error) {
	body, err := s.Request("GET", EndpointChannelMessages(channelID), &MessagesRequest{
		Limit:          limit,
		Before:         beforeTime,
		After:          afterTime,
		IncludePrivate: includePrivate,
	})
	if err != nil {
		return nil, err
	}

	var st []*ChatMessage
	err = unmarshal(body, &st)
	return st, err
}

// ChannelMessageUpdate updates a message in a channel.
// channelID : The ID of a Channel.
// messageID : The ID of a Message.
// data      : The message struct to send.
func (s *Session) ChannelMessageUpdate(channelID, messageID string, data *MessageUpdate) (*ChatMessage, error) {
	body, err := s.Request("PUT", EndpointChannelMessage(channelID, messageID), data)
	if err != nil {
		return nil, err
	}

	var st *ChatMessage
	err = unmarshal(body, &st)
	return st, err
}

// ChannelMessageDelete deletes a message in a channel.
// channelID : The ID of a Channel.
// messageID : The ID of a Message.
func (s *Session) ChannelMessageDelete(channelID, messageID string) error {
	_, err := s.Request("DELETE", EndpointChannelMessage(channelID, messageID), nil)
	return err
}

// ------------------------------------------------------------------------------------------------
// Functions specific to Guilded Servers
// ------------------------------------------------------------------------------------------------

func (s *Session) ServerGet(serverID string) (*Server, error) {
	body, err := s.Request("GET", EndpointServer(serverID), nil)
	if err != nil {
		return nil, err
	}

	var st *Server
	err = unmarshal(body, &st)
	return st, err
}

// ------------------------------------------------------------------------------------------------
// Functions specific to Guilded Channels
// ------------------------------------------------------------------------------------------------

// ChannelCreate creates a channel in a server.
// data : The channel struct to send.
func (s *Session) ChannelCreate(data *ChannelCreate) (*ServerChannel, error) {
	body, err := s.Request("POST", EndpointChannels, data)
	if err != nil {
		return nil, err
	}

	var st *ServerChannel
	err = unmarshal(body, &st)
	return st, err
}

// ChannelGet returns a channel.
// channelID : The ID of a Channel.
func (s *Session) ChannelGet(channelID string) (*ServerChannel, error) {
	body, err := s.Request("GET", EndpointChannel(channelID), nil)
	if err != nil {
		return nil, err
	}

	var st *ServerChannel
	err = unmarshal(body, &st)
	return st, err
}

// ChannelUpdate updates a channel.
// channelID : The ID of a Channel.
// data      : The channel struct to send.
func (s *Session) ChannelUpdate(channelID string, data *ChannelUpdate) (*ServerChannel, error) {
	body, err := s.Request("PATCH", EndpointChannel(channelID), data)
	if err != nil {
		return nil, err
	}

	var st *ServerChannel
	err = unmarshal(body, &st)
	return st, err
}

// ChannelDelete deletes a channel.
// channelID : The ID of a Channel.
func (s *Session) ChannelDelete(channelID string) error {
	_, err := s.Request("DELETE", EndpointChannel(channelID), nil)
	return err
}

// ------------------------------------------------------------------------------------------------
// Functions specific to Guilded Members
// ------------------------------------------------------------------------------------------------

// ServerMemberGet returns a member of the server.
// serverID : The ID of a Server.
// userID   : The ID of a User.
func (s *Session) ServerMemberGet(serverID, userID string) (*ServerMember, error) {
	body, err := s.Request("GET", EndpointServerMember(serverID, userID), nil)
	if err != nil {
		return nil, err
	}

	var st *ServerMember
	err = unmarshal(body, &st)
	return st, err
}

// ServerMemberKick kicks a member from a server.
// serverID : The ID of a Server.
// userID   : The ID of a User.
func (s *Session) ServerMemberKick(serverID, userID string) error {
	_, err := s.Request("DELETE", EndpointServerMember(serverID, userID), nil)
	return err
}

// ServerMembers returns an array of members of a server.
// serverID : The ID of a Server.
func (s *Session) ServerMembers(serverID string) ([]*ServerMember, error) {
	body, err := s.Request("GET", EndpointServerMembers(serverID), nil)
	if err != nil {
		return nil, err
	}

	var st []*ServerMember
	err = unmarshal(body, &st)
	return st, err
}

// ServerMemberNicknameUpdate updates a member's nickname in a server.
// serverID : The ID of a Server.
// userID   : The ID of a User.
// nickname : The nickname to set.
func (s *Session) ServerMemberNicknameUpdate(serverID, userID string, nickname string) (string, error) {
	body, err := s.Request("PUT", EndpointServerMemberNickname(serverID, userID), &ServerMemberNicknameUpdate{
		Nickname: nickname,
	})
	if err != nil {
		return "", err
	}

	var st ServerMemberNicknameUpdate
	err = unmarshal(body, &st)
	return st.Nickname, err
}

// ServerMemberNicknameDelete deletes a member's nickname in a server.
// serverID : The ID of a Server.
// userID   : The ID of a User.
func (s *Session) ServerMemberNicknameDelete(serverID, userID string) error {
	_, err := s.Request("DELETE", EndpointServerMemberNickname(serverID, userID), nil)
	return err
}

// ------------------------------------------------------------------------------------------------
// Functions specific to Guilded Member Bans
// ------------------------------------------------------------------------------------------------

// ServerMemberBanCreate creates a ban on a member of a server.
// serverID : The ID of a Server.
// userID   : The ID of a User.
// reason   : The reason for the ban.
func (s *Session) ServerMemberBanCreate(serverID, userID, reason string) (*ServerMemberBan, error) {
	body, err := s.Request("POST", EndpointServerBansMember(serverID, userID), &ServerMemberBanCreate{
		Reason: reason,
	})
	if err != nil {
		return nil, err
	}

	var st *ServerMemberBan
	err = unmarshal(body, &st)
	return st, err
}

// ServerMemberBan returns a ban on a member of a server.
// serverID : The ID of a Server.
// userID   : The ID of a User.
func (s *Session) ServerMemberBan(serverID, userID string) (*ServerMemberBan, error) {
	body, err := s.Request("GET", EndpointServerBansMember(serverID, userID), nil)
	if err != nil {
		return nil, err
	}

	var st *ServerMemberBan
	err = unmarshal(body, &st)
	return st, err
}

// ServerMemberBanDelete deletes a ban on a member of a server.
// serverID : The ID of a Server.
// userID   : The ID of a User.
func (s *Session) ServerMemberBanDelete(serverID, userID string) error {
	_, err := s.Request("DELETE", EndpointServerBansMember(serverID, userID), nil)
	return err
}

// ServerMemberBans returns an array of bans on a member of a server.
// serverID : The ID of a Server.
func (s *Session) ServerMemberBans(serverID string) ([]*ServerMemberBan, error) {
	body, err := s.Request("GET", EndpointServerBans(serverID), nil)
	if err != nil {
		return nil, err
	}

	var st []*ServerMemberBan
	err = unmarshal(body, &st)
	return st, err
}

// ------------------------------------------------------------------------------------------------
// Functions specific to Guilded Forums
// ------------------------------------------------------------------------------------------------

// ChannelForumTopicCreate creates a topic in a channel.
// channelID : The ID of a Channel.
// title     : The title of the topic.
// content   : The content of the topic.
func (s *Session) ChannelForumTopicCreate(channelID, title, content string) (*ForumTopic, error) {
	body, err := s.Request("POST", EndpointChannelTopics(channelID), &ChannelForumTopicCreate{
		Title:   title,
		Content: content,
	})
	if err != nil {
		return nil, err
	}

	var st *ForumTopic
	err = unmarshal(body, &st)
	return st, err
}

// ------------------------------------------------------------------------------------------------
// Functions specific to Guilded ListItem
// ------------------------------------------------------------------------------------------------

// ChannelListItem creates a list item in a channel.
// channelID : The ID of a Channel.
// data	     : The data for the list item.
func (s *Session) ChannelListItemCreate(channelID string, data *ChannelListItem) (*ListItem, error) {
	body, err := s.Request("POST", EndpointChannelItems(channelID), data)
	if err != nil {
		return nil, err
	}

	var st *ListItem
	err = unmarshal(body, &st)
	return st, err
}

// ChannelListItems returns an array of list items in a channel without notes content.
// channelID : The ID of a Channel.
func (s *Session) ChannelListItems(channelID string) ([]*ListItem, error) {
	body, err := s.Request("GET", EndpointChannelItems(channelID), nil)
	if err != nil {
		return nil, err
	}

	var st []*ListItem
	err = unmarshal(body, &st)
	return st, err
}

// ChannelListItem returns a list item in a channel.
// channelID : The ID of a Channel.
// itemID    : The ID of a ListItem.
func (s *Session) ChannelListItem(channelID, itemID string) (*ListItem, error) {
	body, err := s.Request("GET", EndpointChannelItem(channelID, itemID), nil)
	if err != nil {
		return nil, err
	}

	var st *ListItem
	err = unmarshal(body, &st)
	return st, err
}

// ChannelListItemUpdate updates a list item in a channel.
// channelID : The ID of a Channel.
// itemID    : The ID of a ListItem.
// data	     : The data for the list item.
func (s *Session) ChannelListItemUpdate(channelID, itemID string, data *ChannelListItem) (*ListItem, error) {
	body, err := s.Request("PUT", EndpointChannelItem(channelID, itemID), data)
	if err != nil {
		return nil, err
	}

	var st *ListItem
	err = unmarshal(body, &st)
	return st, err
}

// ChannelListItemDelete deletes a list item in a channel.
// channelID : The ID of a Channel.
// itemID    : The ID of a ListItem.
func (s *Session) ChannelListItemDelete(channelID, itemID string) error {
	_, err := s.Request("DELETE", EndpointChannelItem(channelID, itemID), nil)
	return err
}

// ChannelListItemComplete completes a list item in a channel.
// channelID : The ID of a Channel.
// itemID    : The ID of a ListItem.
func (s *Session) ChannelListItemComplete(channelID, itemID string) error {
	_, err := s.Request("POST", EndpointChannelItemComplete(channelID, itemID), nil)
	return err
}

// ChannelListItemUncomplete uncompletes a list item in a channel.
// channelID : The ID of a Channel.
// itemID    : The ID of a ListItem.
func (s *Session) ChannelListItemUncomplete(channelID, itemID string) error {
	_, err := s.Request("DELETE", EndpointChannelItemComplete(channelID, itemID), nil)
	return err
}

// ------------------------------------------------------------------------------------------------
// Functions specific to Guilded Docs
// ------------------------------------------------------------------------------------------------

// ChannelDocCreate creates a doc in a channel.
// channelID : The ID of a Channel.
// data	     : The data for the doc.
func (s *Session) ChannelDocCreate(channelID string, data *ChannelDoc) (*Doc, error) {
	body, err := s.Request("POST", EndpointChannelDocs(channelID), data)
	if err != nil {
		return nil, err
	}

	var st *Doc
	err = unmarshal(body, &st)
	return st, err
}

// ChannelDoc returns a doc in a channel.
// channelID : The ID of a Channel.
// docID     : The ID of a Doc.
func (s *Session) ChannelDoc(channelID string, docID int) (*Doc, error) {
	body, err := s.Request("GET", EndpointChannelDoc(channelID, fmt.Sprintf("%d", docID)), nil)
	if err != nil {
		return nil, err
	}

	var st *Doc
	err = unmarshal(body, &st)
	return st, err
}

// ChannelDocs returns an array of docs in a channel.
// channelID : The ID of a Channel.
func (s *Session) ChannelDocs(channelID string) ([]*Doc, error) {
	body, err := s.Request("GET", EndpointChannelDocs(channelID), nil)
	if err != nil {
		return nil, err
	}

	var st []*Doc
	err = unmarshal(body, &st)
	return st, err
}

// ChannelDocUpdate updates a doc in a channel.
// channelID : The ID of a Channel.
// docID     : The ID of a Doc.
// data	     : The data for the doc.
func (s *Session) ChannelDocUpdate(channelID string, docID int, data *ChannelDoc) (*Doc, error) {
	body, err := s.Request("PUT", EndpointChannelDoc(channelID, fmt.Sprintf("%d", docID)), data)
	if err != nil {
		return nil, err
	}

	var st *Doc
	err = unmarshal(body, &st)
	return st, err
}

// ChannelDocDelete deletes a doc in a channel.
// channelID : The ID of a Channel.
// docID     : The ID of a Doc.
func (s *Session) ChannelDocDelete(channelID string, docID int) error {
	_, err := s.Request("DELETE", EndpointChannelDoc(channelID, fmt.Sprintf("%d", docID)), nil)
	return err
}

// ------------------------------------------------------------------------------------------------
// Functions specific to Guilded Calendars
// ------------------------------------------------------------------------------------------------

// ChannelEventCreate creates a calendar event in a channel.
// channelID : The ID of a Channel.
// data	     : The data for the event.
func (s *Session) ChannelEventCreate(channelID string, data *ChannelEvent) (*CalendarEvent, error) {
	body, err := s.Request("POST", EndpointChannelEvents(channelID), data)
	if err != nil {
		return nil, err
	}

	var st *CalendarEvent
	err = unmarshal(body, &st)
	return st, err
}

// ChannelEvent returns a calendar event in a channel.
// channelID : The ID of a Channel.
// eventID   : The ID of an CalendarEvent.
func (s *Session) ChannelEvent(channelID string, eventID int) (*CalendarEvent, error) {
	body, err := s.Request("GET", EndpointChannelEvent(channelID, fmt.Sprintf("%d", eventID)), nil)
	if err != nil {
		return nil, err
	}

	var st *CalendarEvent
	err = unmarshal(body, &st)
	return st, err
}

// ChannelEvents returns an array of calendar events in a channel.
// channelID : The ID of a Channel.
// data	     : The data for the event.
func (s *Session) ChannelEvents(channelID string, data *ChannelEventsRequest) ([]*CalendarEvent, error) {
	body, err := s.Request("GET", EndpointChannelEvents(channelID), data)
	if err != nil {
		return nil, err
	}

	var st []*CalendarEvent
	err = unmarshal(body, &st)
	return st, err
}

// ChannelEventUpdate updates a calendar event in a channel.
// channelID : The ID of a Channel.
// eventID   : The ID of an CalendarEvent.
// data	     : The data for the event.
func (s *Session) ChannelEventUpdate(channelID string, eventID int, data *ChannelEvent) (*CalendarEvent, error) {
	body, err := s.Request("PATCH", EndpointChannelEvent(channelID, fmt.Sprintf("%d", eventID)), data)
	if err != nil {
		return nil, err
	}

	var st *CalendarEvent
	err = unmarshal(body, &st)
	return st, err
}

// ChannelEventDelete deletes a calendar event in a channel.
// channelID : The ID of a Channel.
// eventID   : The ID of an CalendarEvent.
func (s *Session) ChannelEventDelete(channelID string, eventID int) error {
	_, err := s.Request("DELETE", EndpointChannelEvent(channelID, fmt.Sprintf("%d", eventID)), nil)
	return err
}

// ------------------------------------------------------------------------------------------------
// Functions specific to Guilded Reactions
// ------------------------------------------------------------------------------------------------

// ChannelContentReactionAdd adds a reaction to a content in a channel.
// channelID : The ID of a Channel.
// contentID : The ID of a Content.
// emoteID   : The ID of an Emote.
func (s *Session) ChannelContentReactionAdd(channelID, contentID string, emoteID int) error {
	_, err := s.Request("PUT", EndpointChannelReaction(channelID, contentID, fmt.Sprintf("%d", emoteID)), nil)
	return err
}

// ChannelContentReactionDelete deletes a reaction to a content in a channel.
// channelID : The ID of a Channel.
// contentID : The ID of a Content.
// emoteID   : The ID of an Emote.
func (s *Session) ChannelContentReactionDelete(channelID, contentID string, emoteID int) error {
	_, err := s.Request("DELETE", EndpointChannelReaction(channelID, contentID, fmt.Sprintf("%d", emoteID)), nil)
	return err
}

// ------------------------------------------------------------------------------------------------
// Functions specific to Guilded XP
// ------------------------------------------------------------------------------------------------

// ServerMemberXPAward awards XP to a member of a server.
// serverID : The ID of a Server.
// memberID : The ID of a Member.
// amount   : The amount of XP to award.
func (s *Session) ServerMemberXPAward(serverID, memberID string, amount int) (int, error) {
	body, err := s.Request("POST", EndpointServerXPMember(serverID, memberID), &ServerXPUpdate{Amount: amount})
	if err != nil {
		return 0, err
	}

	var st struct {
		Amount int `json:"amount"`
	}
	err = unmarshal(body, &st)
	return st.Amount, err
}

// ServerRoleXPAward awards XP to a role of a server.
// serverID : The ID of a Server.
// roleID   : The ID of a Role.
// amount   : The amount of XP to award.
func (s *Session) ServerRoleXPAward(serverID, roleID string, amount int) (int, error) {
	body, err := s.Request("POST", EndpointServerXPRoles(serverID, roleID), &ServerXPUpdate{Amount: amount})
	if err != nil {
		return 0, err
	}

	var st struct {
		Amount int `json:"amount"`
	}
	err = unmarshal(body, &st)
	return st.Amount, err
}

// ------------------------------------------------------------------------------------------------
// Functions specific to Guilded Social Links
// ------------------------------------------------------------------------------------------------

// ServerMemberSocialLink returns a social link of a member of a server.
// serverID : The ID of a Server.
// memberID : The ID of a Member.
// linkType : The type of the social-link.
func (s *Session) ServerMemberSocialLink(serverID, memberID, linkType string) (*ServerSocialLink, error) {
	body, err := s.Request("GET", EndpointServerMemberSocialLink(serverID, memberID, linkType), nil)
	if err != nil {
		return nil, err
	}

	var st *ServerSocialLink
	err = unmarshal(body, &st)
	return st, err
}

// ------------------------------------------------------------------------------------------------
// Functions specific to Guilded Group Membership
// ------------------------------------------------------------------------------------------------

// GroupMemberAdd adds a member to a group.
// groupID : The ID of a Group.
// memberID : The ID of a Member.
func (s *Session) GroupMemberAdd(groupID, userID string) error {
	_, err := s.Request("PUT", EndpointGroupMember(groupID, userID), nil)
	return err
}

// GroupMemberRemove removes a member from a group.
// groupID : The ID of a Group.
// memberID : The ID of a Member.
func (s *Session) GroupMemberRemove(groupID, userID string) error {
	_, err := s.Request("DELETE", EndpointGroupMember(groupID, userID), nil)
	return err
}

// ------------------------------------------------------------------------------------------------
// Functions specific to Guilded Role Membership
// ------------------------------------------------------------------------------------------------

// ServerMemberRoleAdd adds a role to a member of a server.
// serverID : The ID of a Server.
// memberID : The ID of a Member.
// roleID   : The ID of a Role.
func (s *Session) ServerMemberRoleAdd(serverID, memberID string, roleID int) error {
	_, err := s.Request("PUT", EndpointServerMemberRole(serverID, memberID, fmt.Sprintf("%d", roleID)), nil)
	return err
}

// ServerMemberRoleRemove removes a role from a member of a server.
// serverID : The ID of a Server.
// memberID : The ID of a Member.
// roleID   : The ID of a Role.
func (s *Session) ServerMemberRoleRemove(serverID, memberID string, roleID int) error {
	_, err := s.Request("DELETE", EndpointServerMemberRole(serverID, memberID, fmt.Sprintf("%d", roleID)), nil)
	return err
}

// ServerMemberRoles returns a list of roles of a member of a server.
// serverID : The ID of a Server.
// memberID : The ID of a Member.
func (s *Session) ServerMemberRoles(serverID, memberID string) ([]int, error) {
	body, err := s.Request("GET", EndpointServerMemberRoles(serverID, memberID), nil)
	if err != nil {
		return nil, err
	}

	var st struct {
		RoleIDs []int `json:"roleIds"`
	}
	err = unmarshal(body, &st)
	return st.RoleIDs, err
}

// ------------------------------------------------------------------------------------------------
// Functions specific to Guilded Webhooks
// ------------------------------------------------------------------------------------------------

// ServerWebhookCreate creates a webhook in a server.
// serverID : The ID of a Server.
// data	    : The data for the webhook.
func (s *Session) ServerWebhookCreate(serverID string, data *WebhookCreate) (*Webhook, error) {
	body, err := s.Request("POST", EndpointServerWeebhooks(serverID), data)
	if err != nil {
		return nil, err
	}

	var st *Webhook
	err = unmarshal(body, &st)
	return st, err
}

// ServerWebhook returns a webhook in a server.
// serverID  : The ID of a Server.
// webhookID : The ID of a Webhook.
func (s *Session) ServerWebhook(serverID, webhookID string) (*Webhook, error) {
	body, err := s.Request("GET", EndpointServerWeebhook(serverID, webhookID), nil)
	if err != nil {
		return nil, err
	}

	var st *Webhook
	err = unmarshal(body, &st)
	return st, err
}

// ServerWebhooks returns a list of webhooks in a server.
// serverID : The ID of a Server.
func (s *Session) ServerWebhooks(serverID string) ([]*Webhook, error) {
	body, err := s.Request("GET", EndpointServerWeebhooks(serverID), nil)
	if err != nil {
		return nil, err
	}

	var st struct {
		Webhooks []*Webhook `json:"webhooks"`
	}
	err = unmarshal(body, &st)
	return st.Webhooks, err
}

// ServerWebhookUpdate updates a webhook in a server.
// serverID  : The ID of a Server.
// webhookID : The ID of a Webhook.
// data	     : The data for the webhook.
func (s *Session) ServerWebhookUpdate(serverID, webhookID string, data *WebhookUpdate) (*Webhook, error) {
	body, err := s.Request("PUT", EndpointServerWeebhook(serverID, webhookID), data)
	if err != nil {
		return nil, err
	}

	var st *Webhook
	err = unmarshal(body, &st)
	return st, err
}

func (s *Session) ServerWebhookDelete(serverID, webhookID string) error {
	_, err := s.Request("DELETE", EndpointServerWeebhook(serverID, webhookID), nil)
	return err
}
