# Microsoft Teams 3rd Party Device API Library in Golang

I reverse engineered the Microsoft Teams 3rd Party Device API and wrote this golang library to interact with it.

Check out the example for how it all works. It is pretty simple code.

## Basics

You have to have enabled the 3rd party device API on your MS Teams application prior to trying to call it

### Create a client

Everything revolves around a `Client` object that is created by `func NewClient(manufacturer, device, app, appVersion, token string, port int) *Client`

You can come up with whatever strings you want for manufacturer, device, and app. I've been providing version as `X.Y.Z` format.

Until you actually have a `token`, you can pass that as an empty string.

The port is usually `8124`, and will default to that if you pass a `0`

### Connect

`client.Connect()` begins a session with MS Teams and will auto-reconnect until you call `client.Disconnect()`

It will not create a token until you try an action or reaction like muting, raising your hand, etc. At that point, MS Teams will display a dialog asking the user if they want to allow your app to control it. If they do, you will receive a token.

### Token Management

Once the client has a token, you can get it and store it with `client.GetToken()`. The token may change during your session, so it is best practice to get the current token from the client before shutdown. As long as you keep track of this, you won't have to display a permissions dialog in MS Teams everytime your app restarts.

### Actions

Look through the public methods in `client.go`

Common ones are

`func (c *Client) ToggleMute()`

`func (c *Client) ToggleVideo()`

### Keeping track of state

There are convenience functions you may use like 

`func (c *Client) CanToggleMute() bool`

But if you register an event callback with `func (c *Client) SetEventCallback(callback EventCallback)`, then you will get updates anytime something changes. 


