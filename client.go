package TeamsClientDeviceLibrary

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"sync"
	"time"
)

type Client struct {
	messageId      int
	messageIdMutex sync.Mutex

	isConnected      bool
	isConnectedMutex sync.Mutex

	manufacturer   string
	device         string
	app            string
	appVersion     string
	port           int
	autoReconnect  bool
	callback       EventCallback
	errorCallback  ErrorCallback
	disconnectChan chan int

	tokenMutex sync.RWMutex
	token      string

	lastUpdateMutex sync.RWMutex
	lastUpdate      TeamsMeetingUpdate

	sendChan chan TeamsOutboundMessage
}

type EventCallback func(update TeamsMeetingUpdate)
type ErrorCallback func(err error)

/*
If you don't have a token, pass a blank string. It will prompt you to approve in teams and then give you one.
You should always check the token before quitting the program
*/
func NewClient(manufacturer, device, app, appVersion, token string, port int) *Client {
	if port == 0 {
		port = 8124
	}
	return &Client{
		manufacturer:     manufacturer,
		device:           device,
		app:              app,
		appVersion:       appVersion,
		port:             port,
		token:            token,
		tokenMutex:       sync.RWMutex{},
		lastUpdateMutex:  sync.RWMutex{},
		autoReconnect:    true,
		messageId:        1,
		messageIdMutex:   sync.Mutex{},
		disconnectChan:   make(chan int),
		sendChan:         make(chan TeamsOutboundMessage),
		isConnectedMutex: sync.Mutex{},
	}
}

func (c *Client) Connect() error {
	tokenStr := ""
	if c.token != "" {
		tokenStr = fmt.Sprintf("&token=%s", c.token)
	}
	u := url.URL{
		Scheme:   "ws",
		Host:     fmt.Sprintf("127.0.0.1:%d", c.port),
		Path:     "/",
		RawQuery: fmt.Sprintf("protocol-version=2.0.0&manufacturer=%s&device=%s&app=%s&app-version=%s%s", c.manufacturer, c.device, c.app, c.appVersion, tokenStr),
	}

	c.autoReconnect = true

	msgChannels := newWsChannelsWithSender(c.sendChan)
	ctx, cancelCtxFunc := context.WithCancel(context.Background())

	go func() {
		defer close(msgChannels.receiveChan)
		for {
			select {
			case msg := <-msgChannels.receiveChan:
				c.handleMessage(msg)
			case <-c.disconnectChan:
				cancelCtxFunc()
				return
			}
		}
	}()

	go func() {
		for {
			//TODO be less lazy about this
			c.isConnectedMutex.Lock()
			c.isConnected = true
			c.isConnectedMutex.Unlock()
			wsLoop(u.String(), msgChannels, ctx)
			c.isConnectedMutex.Lock()
			c.isConnected = false
			c.isConnectedMutex.Unlock()
			if !c.autoReconnect { //this is disabled with disconnect is called
				return
			}
			time.Sleep(500 * time.Millisecond)
		}
	}()

	return nil
}

func (c *Client) getNextMessageId() int {
	c.messageIdMutex.Lock()
	defer c.messageIdMutex.Unlock()
	c.messageId += 1
	return c.messageId
}

func (c *Client) IsConnected() bool {
	c.isConnectedMutex.Lock()
	defer c.isConnectedMutex.Unlock()
	return c.isConnected
}

func (c *Client) Disconnect() error {
	c.autoReconnect = false
	if c.isConnected {
		c.disconnectChan <- 0
	}
	//TODO nope, that's not right
	return fmt.Errorf("not connected")
}

func (c *Client) handleMessage(msg TeamsIncomingMessage) {
	if msg.TokenRefresh != "" {
		c.token = msg.TokenRefresh
		log.Printf("New token: %s", c.token)
	} else if msg.Response == "Success" {
		log.Printf("Teams successfully acknowledged message: %d", msg.RequestId)
	} else {
		if c.callback != nil {
			c.callback(msg.MeetingUpdate)
		}
	}
}

func (c *Client) SetEventCallback(callback EventCallback) {
	c.callback = callback
}

func (c *Client) SetErrorCallback(callback ErrorCallback) {
	c.errorCallback = callback
}

func (c *Client) GetToken() string {
	c.tokenMutex.Lock()
	defer c.tokenMutex.Unlock()
	return c.token
}

func (c *Client) SetToken(token string) {
	c.tokenMutex.Lock()
	defer c.tokenMutex.Unlock()
	c.token = token
}

func (c *Client) CanToggleMute() bool {
	c.lastUpdateMutex.Lock()
	defer c.lastUpdateMutex.Unlock()
	return c.lastUpdate.MeetingPermissions.CanToggleMute
}

func (c *Client) IsMuted() bool {
	c.lastUpdateMutex.Lock()
	defer c.lastUpdateMutex.Unlock()
	return c.lastUpdate.MeetingState.IsMuted
}

func (c *Client) IsVideoOn() bool {
	c.lastUpdateMutex.Lock()
	defer c.lastUpdateMutex.Unlock()
	return c.lastUpdate.MeetingState.IsVideoOn
}

func (c *Client) CanToggleVideo() bool {
	c.lastUpdateMutex.Lock()
	defer c.lastUpdateMutex.Unlock()
	return c.lastUpdate.MeetingPermissions.CanToggleVideo
}

func (c *Client) IsHandRaised() bool {
	c.lastUpdateMutex.Lock()
	defer c.lastUpdateMutex.Unlock()
	return c.lastUpdate.MeetingState.IsHandRaised
}

func (c *Client) CanRaiseHand() bool {
	c.lastUpdateMutex.Lock()
	defer c.lastUpdateMutex.Unlock()
	return c.lastUpdate.MeetingPermissions.CanToggleHand
}

func (c *Client) CanLeave() bool {
	c.lastUpdateMutex.Lock()
	defer c.lastUpdateMutex.Unlock()
	return c.lastUpdate.MeetingPermissions.CanLeave
}

func (c *Client) outboundAction(action TeamsAction, modifier TeamsActionModifier) {
	msg := TeamsOutboundMessage{
		Action:    string(action),
		RequestId: c.getNextMessageId(),
	}
	if modifier != ModifierNone {
		msg.Parameters = []byte(fmt.Sprintf("{\"type\":\"%s\"}", modifier))
	}
	c.sendChan <- msg
}

func (c *Client) Refresh() {
	c.outboundAction(RefreshState, ModifierNone)
}

func (c *Client) ToggleMute() {
	c.outboundAction(ToggleMute, ModifierNone)
}

func (c *Client) ToggleVideo() {
	c.outboundAction(ToggleVideo, ModifierNone)
}

func (c *Client) ToggleVideoBackgroundBlur() {
	c.outboundAction(ToggleBackgroundBlur, ModifierNone)
}

func (c *Client) ToggleHandRaised() {
	c.outboundAction(ToggleHand, ModifierNone)
}

func (c *Client) ToggleShareTrap() {
	c.outboundAction(ToggleUIWithParams, ToggleUIShareTray)
}

func (c *Client) ShopSharing() {
	c.outboundAction(StopSharing, ModifierNone)
}

func (c *Client) ToggleChat() {
	c.outboundAction(ToggleUIWithParams, ToggleUIChat)
}

func (c *Client) Leave() {
	c.outboundAction(Leave, ModifierNone)
}

func (c *Client) ReactLove() {
	c.outboundAction(React, ReactionLove)
}

func (c *Client) ReactLaugh() {
	c.outboundAction(React, ReactionLaugh)
}

func (c *Client) ReactApplause() {
	c.outboundAction(React, ReactionApplause)
}

func (c *Client) ReactWow() {
	c.outboundAction(React, ReactionWow)
}

func (c *Client) ReactLike() {
	c.outboundAction(React, ReactionLike)
}
