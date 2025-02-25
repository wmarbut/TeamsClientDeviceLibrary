package TeamsClientDeviceLibrary

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/websocket"
)

type wsChannels struct {
	sendChan    chan TeamsOutboundMessage //the channel for the program to send messages to MS Teams
	receiveChan chan TeamsIncomingMessage //the channel for the program to receive messages from MS Teams
}

func newWsChannelsWithSender(sender chan TeamsOutboundMessage) wsChannels {
	return wsChannels{
		sendChan:    sender,
		receiveChan: make(chan TeamsIncomingMessage),
	}
}

func newWsChannels() wsChannels {
	return newWsChannelsWithSender(make(chan TeamsOutboundMessage))
}

func wsLoop(teamsUrl string, msgChannels wsChannels, ctx context.Context) error {
	//Wrap the parent context with another cancellable context
	//We will trigger the child cancel when/if the parent cancels
	wsContext, wsContextCancel := context.WithCancel(ctx)
	websocketConn, _, err := websocket.DefaultDialer.DialContext(wsContext, teamsUrl, nil)
	if err != nil {
		log.Printf("Dial error: %s for url: %s", err, teamsUrl)
		return err
	}
	defer websocketConn.Close()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	readDisconnectChan := make(chan int, 1)
	go func() {
		for {
			_, message, err := websocketConn.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err) {
					log.Printf("Exiting read loop due to closed connection")
					readDisconnectChan <- 1
				} else {
					log.Printf("Unable to read message, disconnecting with error: %s\n", err)
					readDisconnectChan <- 1
				}
				return
			}
			msg := TeamsIncomingMessage{}
			err = json.Unmarshal(message, &msg)
			//log.Printf("Raw message: \"%s\"", string(message))
			if err != nil {
				log.Printf("Unable to unmarshal message: %s\n", message)
				log.Printf("Error: %s\n", err)
				continue
			}
			msgChannels.receiveChan <- msg
		}
	}()

	go func() {
		for {
			select {
			case <-wsContext.Done():
				log.Printf("Leaving WS write loop due to context cancellation")
				return
			case toSend := <-msgChannels.sendChan:
				jsonStr, err := json.Marshal(toSend)
				if err != nil {
					log.Printf("Error marshalling teams command: %s\n", err)
					continue
				}
				err = websocketConn.WriteMessage(websocket.TextMessage, jsonStr)
				if err != nil {
					log.Printf("Error sending teams command: %s\n", err)
					continue
				}
				log.Printf("Sent command %d of type: %s", toSend.RequestId, toSend.Action)
			}
		}
	}()

	select {
	case <-sigChan:
		log.Printf("Received os exit signal\n")
		wsContextCancel()

	case <-ctx.Done():
		log.Printf("wsLoop received parent context cancellation")
		wsContextCancel()

	case <-readDisconnectChan:
		log.Printf("wsLoop read loop signalled a disconnect")
		wsContextCancel()

	}
	log.Printf("Leaving wsLoop")
	return nil
}
