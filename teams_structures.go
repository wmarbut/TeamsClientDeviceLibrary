package TeamsClientDeviceLibrary

import "encoding/json"

type TeamsAction string

const (
	ToggleMute           TeamsAction = "toggle-mute"
	ToggleVideo          TeamsAction = "toggle-video"
	ToggleHand           TeamsAction = "toggle-hand"
	Leave                TeamsAction = "leave-call"
	RefreshState         TeamsAction = "query-state"
	ToggleBackgroundBlur TeamsAction = "toggle-background-blur"
	ToggleUIWithParams   TeamsAction = "toggle-ui"
	StopSharing          TeamsAction = "stop-sharing"
	React                TeamsAction = "send-reaction"
)

type TeamsActionModifier string

const (
	ToggleUIChat      TeamsActionModifier = "chat"
	ToggleUIShareTray TeamsActionModifier = "share-tray"
	ReactionApplause  TeamsActionModifier = "applause"
	ReactionLove      TeamsActionModifier = "love"
	ReactionLaugh     TeamsActionModifier = "laugh"
	ReactionWow       TeamsActionModifier = "wow"
	ReactionLike      TeamsActionModifier = "like"
	ModifierNone      TeamsActionModifier = ""
)

type TeamsIncomingMessage struct {
	TokenRefresh  string             `json:"tokenRefresh"`  //only present with a new token
	MeetingUpdate TeamsMeetingUpdate `json:"meetingUpdate"` //status updates only
	RequestId     int                `json:"requestId"`     //used in confirmation messages
	Response      string             `json:"response"`      //used in confirmation messages
}
type TeamsMeetingPermissions struct {
	CanToggleMute      bool `json:"canToggleMute"`
	CanToggleVideo     bool `json:"canToggleVideo"`
	CanToggleHand      bool `json:"canToggleHand"`
	CanToggleBlur      bool `json:"canToggleBlur"`
	CanLeave           bool `json:"canLeave"`
	CanReact           bool `json:"canReact"`
	CanToggleShareTray bool `json:"canToggleShareTray"`
	CanToggleChat      bool `json:"canToggleChat"`
	CanStopSharing     bool `json:"canStopSharing"`
	CanPair            bool `json:"canPair"`
}

type TeamsMeetingState struct {
	IsMuted             bool `json:"isMuted"`
	IsVideoOn           bool `json:"isVideoOn"`
	IsHandRaised        bool `json:"isHandRaised"`
	IsInMeeting         bool `json:"isInMeeting"`
	IsRecordingOn       bool `json:"isRecordingOn"`
	IsBackgroundBlurred bool `json:"isBackgroundBlurred"`
	IsSharing           bool `json:"isSharing"`
	HasUnreadMessages   bool `json:"hasUnreadMessages"`
}

type TeamsMeetingUpdate struct {
	MeetingState       TeamsMeetingState       `json:"meetingState"`
	MeetingPermissions TeamsMeetingPermissions `json:"meetingPermissions"`
}

type TeamsOutboundMessage struct {
	Action     string          `json:"action"`
	Parameters json.RawMessage `json:"parameters"`
	RequestId  int             `json:"requestId"`
}
