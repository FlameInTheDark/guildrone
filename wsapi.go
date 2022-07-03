package guildrone

import (
	"bytes"
	"compress/zlib"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"net/http"
	"time"
)

// ErrWSAlreadyOpen is thrown when you attempt to open
// a websocket that already is open.
var ErrWSAlreadyOpen = errors.New("web socket already opened")

// ErrWSNotFound is thrown when you attempt to use a websocket
// that doesn't exist
var ErrWSNotFound = errors.New("no websocket connection exists")

type helloOp struct {
	HeartbeatIntervalMs time.Duration `json:"heartbeatIntervalMs"`
}

// Open creates a websocket connection to Guilded.
// See: https://www.guilded.gg/docs/api/connecting
func (s *Session) Open() error {
	s.log(LogInformational, "called")

	var err error

	// Prevent Open or other major Session functions from
	// being called while Open is still running.
	s.Lock()
	defer s.Unlock()

	// If the websock is already open, bail out here.
	if s.wsConn != nil {
		return ErrWSAlreadyOpen
	}

	// Connect to the Gateway
	s.log(LogInformational, "connecting to gateway %s", EndpointGuildedWebsocket)
	header := http.Header{}
	header.Add("accept-encoding", "zlib")
	header.Add("Authorization", fmt.Sprintf("Bearer %s", s.Token))
	s.wsConn, _, err = websocket.DefaultDialer.Dial(EndpointGuildedWebsocket, header)
	if err != nil {
		s.log(LogError, "error connecting to gateway %s, %s", EndpointGuildedWebsocket, err)
		s.wsConn = nil // Just to be safe.
		return err
	}

	s.wsConn.SetCloseHandler(func(code int, text string) error {
		return nil
	})

	defer func() {
		// because of this, all code below must set err to the error
		// when exiting with an error :)  Maybe someone has a better
		// way :)
		if err != nil {
			s.wsConn.Close()
			s.wsConn = nil
		}
	}()

	// The first response from Guilded should be an Op 1 (Hello) Packet.
	// When processed by onEvent the heartbeat goroutine will be started.
	mt, m, err := s.wsConn.ReadMessage()
	if err != nil {
		return err
	}
	e, err := s.onEvent(mt, m)
	if err != nil {
		return err
	}
	if e.Operation != 1 {
		err = fmt.Errorf("expecting Op 1, got Op %d instead", e.Operation)
		return err
	}
	s.log(LogInformational, "Op 1 Hello Packet received from Guilded")
	var h helloOp
	if err = json.Unmarshal(e.RawData, &h); err != nil {
		err = fmt.Errorf("error unmarshalling helloOp, %s", err)
		return err
	}
	//
	//// Now we send either an Op 2 Identity if this is a brand new
	//// connection or Op 6 Resume if we are resuming an existing connection.
	//sequence := atomic.LoadInt64(s.sequence)
	//if s.sessionID == "" && sequence == 0 {
	//
	//	// Send Op 2 Identity Packet
	//	err = s.identify()
	//	if err != nil {
	//		err = fmt.Errorf("error sending identify packet to gateway, %s, %s", s.gateway, err)
	//		return err
	//	}
	//
	//} else {
	//
	//	// Send Op 6 Resume Packet
	//	p := resumePacket{}
	//	p.Op = 6
	//	p.Data.Token = s.Token
	//	p.Data.SessionID = s.sessionID
	//	p.Data.Sequence = sequence
	//
	//	s.log(LogInformational, "sending resume packet to gateway")
	//	s.wsMutex.Lock()
	//	err = s.wsConn.WriteJSON(p)
	//	s.wsMutex.Unlock()
	//	if err != nil {
	//		err = fmt.Errorf("error sending gateway resume packet, %s, %s", s.gateway, err)
	//		return err
	//	}
	//
	//}

	s.handleEvent(connectEventType, &Connect{})

	// Create listening chan outside of listen, as it needs to happen inside the
	// mutex lock and needs to exist before calling heartbeat and listen
	// go rountines.
	s.listening = make(chan interface{})

	// Start sending heartbeats and reading messages from Guilded.
	go s.heartbeat(s.wsConn, s.listening, h.HeartbeatIntervalMs)
	go s.listen(s.wsConn, s.listening)

	s.log(LogInformational, "exiting")
	return nil
}

// onEvent is the "event handler" for all messages received on the
// Guilded Gateway API websocket connection.
//
// If you use the AddHandler() function to register a handler for a
// specific event this function will pass the event along to that handler.
//
// If you use the AddHandler() function to register a handler for the
// "OnEvent" event then all events will be passed to that handler.
func (s *Session) onEvent(messageType int, message []byte) (*Event, error) {

	var err error
	var reader io.Reader
	reader = bytes.NewBuffer(message)

	// If this is a compressed message, uncompress it.
	if messageType == websocket.BinaryMessage {

		z, err2 := zlib.NewReader(reader)
		if err2 != nil {
			s.log(LogError, "error uncompressing websocket message, %s", err)
			return nil, err2
		}

		defer func() {
			err3 := z.Close()
			if err3 != nil {
				s.log(LogWarning, "error closing zlib, %s", err)
			}
		}()

		reader = z
	}

	if messageType == websocket.PongMessage {
		s.Lock()
		s.LastHeartbeatAck = time.Now().UTC()
		s.Unlock()
		s.log(LogDebug, "got heartbeat ACK")
		return nil, nil
	}

	// Decode the event into an Event struct.
	var e *Event
	decoder := json.NewDecoder(reader)
	if err = decoder.Decode(&e); err != nil {
		s.log(LogError, "error decoding websocket message, %s", err)
		return e, err
	}

	s.log(LogDebug, "Op: %d, MsgID: %d, Type: %s, Data: %s\n\n", e.Operation, e.MessageID, e.Type, string(e.RawData))

	if e.Operation == 1 {
		// Op1 is handled by Open()
		return e, nil
	}

	// Map event to registered event handlers and pass it along to any registered handlers.
	if eh, ok := registeredInterfaceProviders[e.Type]; ok {
		e.Struct = eh.New()

		// Attempt to unmarshal our event.
		if err = json.Unmarshal(e.RawData, e.Struct); err != nil {
			s.log(LogError, "error unmarshalling %s event, %s", e.Type, err)
		}

		s.handleEvent(e.Type, e.Struct)
	} else {
		s.log(LogWarning, "unknown event: Op: %d, MsgID: %d, Type: %s, Data: %s", e.Operation, e.MessageID, e.Type, string(e.RawData))
	}

	// For legacy reasons, we send the raw event also, this could be useful for handling unknown events.
	s.handleEvent(eventEventType, e)

	return e, nil
}

// listen polls the websocket connection for events, it will stop when the
// listening channel is closed, or an error occurs.
func (s *Session) listen(wsConn *websocket.Conn, listening <-chan interface{}) {

	s.log(LogInformational, "called")

	for {

		messageType, message, err := wsConn.ReadMessage()

		if err != nil {

			// Detect if we have been closed manually. If a Close() has already
			// happened, the websocket we are listening on will be different to
			// the current session.
			s.RLock()
			sameConnection := s.wsConn == wsConn
			s.RUnlock()

			if sameConnection {

				s.log(LogWarning, "error reading from gateway %s websocket, %s", EndpointGuildedWebsocket, err)
				// There has been an error reading, close the websocket so that
				// OnDisconnect event is emitted.
				err := s.Close()
				if err != nil {
					s.log(LogWarning, "error closing session connection, %s", err)
				}

				s.log(LogInformational, "calling reconnect() now")
				s.reconnect()
			}

			return
		}

		select {

		case <-listening:
			return

		default:
			s.onEvent(messageType, message)
		}
	}
}

// FailedHeartbeatAcks is the Number of heartbeat intervals to wait until forcing a connection restart.
const FailedHeartbeatAcks time.Duration = 5 * time.Millisecond

// heartbeat sends regular heartbeats to Guilded so it knows the client
// is still connected.  If you do not send these heartbeats Guilded will
// disconnect the websocket connection after a few seconds.
func (s *Session) heartbeat(wsConn *websocket.Conn, listening <-chan interface{}, heartbeatIntervalMsec time.Duration) {

	wsConn.SetPongHandler(func(string) error {
		s.Lock()
		s.LastHeartbeatAck = time.Now().UTC()
		s.Unlock()
		s.log(LogDebug, "got heartbeat ACK")
		return nil
	})

	s.log(LogInformational, "called")

	if listening == nil || wsConn == nil {
		return
	}

	var err error
	ticker := time.NewTicker(heartbeatIntervalMsec * time.Millisecond)
	defer ticker.Stop()

	for {
		s.RLock()
		last := s.LastHeartbeatAck
		s.RUnlock()
		s.log(LogDebug, "sending gateway websocket heartbeat seq")
		s.wsMutex.Lock()
		s.LastHeartbeatSent = time.Now().UTC()
		err = wsConn.WriteMessage(websocket.PingMessage, []byte{})
		s.wsMutex.Unlock()
		if err != nil || time.Now().UTC().Sub(last) > (heartbeatIntervalMsec*FailedHeartbeatAcks) {
			if err != nil {
				s.log(LogError, "error sending heartbeat to gateway %s, %s", EndpointGuildedWebsocket, err)
			} else {
				s.log(LogError, "haven't gotten a heartbeat ACK in %v, triggering a reconnection", time.Now().UTC().Sub(last))
			}
			s.Close()
			s.reconnect()
			return
		}
		s.Lock()
		s.DataReady = true
		s.Unlock()

		select {
		case <-ticker.C:
			// continue loop and send heartbeat
		case <-listening:
			return
		}
	}
}

// Close closes a websocket and stops all listening/heartbeat goroutines.
func (s *Session) Close() error {
	return s.CloseWithCode(websocket.CloseNormalClosure)
}

// CloseWithCode closes a websocket using the provided closeCode and stops all
// listening/heartbeat goroutines.
func (s *Session) CloseWithCode(closeCode int) (err error) {

	s.log(LogInformational, "called")
	s.Lock()

	s.DataReady = false

	if s.listening != nil {
		s.log(LogInformational, "closing listening channel")
		close(s.listening)
		s.listening = nil
	}

	// TODO: Close all active Voice Connections too
	// this should force stop any reconnecting voice channels too

	if s.wsConn != nil {

		s.log(LogInformational, "sending close frame")
		// To cleanly close a connection, a client should send a close
		// frame and wait for the server to close the connection.
		s.wsMutex.Lock()
		err := s.wsConn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(closeCode, ""))
		s.wsMutex.Unlock()
		if err != nil {
			s.log(LogInformational, "error closing websocket, %s", err)
		}

		// TODO: Wait for Guilded to actually close the connection.
		time.Sleep(1 * time.Second)

		s.log(LogInformational, "closing gateway websocket")
		err = s.wsConn.Close()
		if err != nil {
			s.log(LogInformational, "error closing websocket, %s", err)
		}

		s.wsConn = nil
	}

	s.Unlock()

	s.log(LogInformational, "emit disconnect event")
	s.handleEvent(disconnectEventType, &Disconnect{})

	return
}

func (s *Session) reconnect() {

	s.log(LogInformational, "called")

	var err error

	if s.ShouldReconnectOnError {

		wait := time.Duration(1)

		for {
			s.log(LogInformational, "trying to reconnect to gateway")

			err = s.Open()
			if err == nil {
				s.log(LogInformational, "successfully reconnected to gateway")
				return
			}

			// Certain race conditions can call reconnect() twice. If this happens, we
			// just break out of the reconnect loop
			if err == ErrWSAlreadyOpen {
				s.log(LogInformational, "Websocket already exists, no need to reconnect")
				return
			}

			s.log(LogError, "error reconnecting to gateway, %s", err)

			<-time.After(wait * time.Second)
			wait *= 2
			if wait > 600 {
				wait = 600
			}
		}
	}
}
