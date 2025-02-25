package main

import (
	"context"
	"fmt"
	"log"
	"github.com/wmarbut/TeamsClientDeviceLibrary"
	"os"
	"strconv"
	"time"

	"github.com/charmbracelet/huh"
)

const (
	Mute      = "Mute"
	HandRaise = "Raise Hand"
)

type ChoiceSelection struct {
	action string
}

func main() {
	/*Setup the Teams Client. Does not automatically connect*/
	client := TeamsClientDeviceLibrary.NewClient("Dev-Example", "Mac%20Test-Example", "Mac%20Test-Example", "1.0.0", "", 0)

	/*Call to connect to Teams. Will reconnect automatically if dropped until you tell it to Disconnect() */
	client.Connect()

	/* Register an event callback to let you know when something happend.
	 * In this case, we're going to create some text to display
	 */
	var lastMessageText string
	msgHandler := func(msg TeamsClientDeviceLibrary.TeamsMeetingUpdate) {
		lastMessageText = fmt.Sprintf("Mutable %v\nIs Muted %v\nIs Hand Raised %v\n",
			msg.MeetingPermissions.CanToggleMute,
			msg.MeetingState.IsMuted,
			msg.MeetingState.IsHandRaised)
	}
	client.SetEventCallback(msgHandler)

	/* Create an ability for us to interrupt the program loop and exit if we want */
	ctx, ctxCancel := context.WithCancel(context.Background())

	/* Fancy TUI stuff :) */
	go func() {
		choiceSelection := ChoiceSelection{}
		accessible, _ := strconv.ParseBool(os.Getenv("ACCESSIBLE"))

		for {

			form := huh.NewForm(
				huh.NewGroup(
					huh.NewNote().Description(lastMessageText),
					huh.NewSelect[string]().
						Options(huh.NewOptions("Mute", "Raise Hand", "Quit")...).
						Title("Act").
						Validate(func(t string) error {
							return nil
						}).
						Value(&choiceSelection.action),
				),
			).WithAccessible(accessible)
			err := form.Run()
			if err != nil {
				log.Printf("Error running form: %s", err)
			}

			/* Ok, so we have an action from the TUI, let's ask teams to do something about it.
			 * We will get a message back that our event handler can use
			 */
			log.Printf("Performing action: %s", choiceSelection.action)
			if choiceSelection.action == Mute {
				client.ToggleMute()
			} else if choiceSelection.action == HandRaise {
				client.ToggleHandRaised()
			} else if choiceSelection.action == "Quit" {
				ctxCancel()
				break
			}
			log.Printf("Ready to form again")
		}
	}()

	<-ctx.Done()
	client.Disconnect()
	time.Sleep(2 * time.Second)
	log.Println("Exiting...")
}
