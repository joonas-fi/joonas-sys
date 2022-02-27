package statusbar

// i3bar protocol, to allow sending bar items to i3bar

import (
	"encoding/json"
	"fmt"
)

type barItem struct {
	Name     string `json:"name"`
	Instance string `json:"instance"`
	Markup   string `json:"markup"`
	FullText string `json:"full_text"`
	Color    string `json:"color,omitempty"` // example: "#FF0000"
}

type i3barProtocolSender struct{}

// also calls sendHeaders() because without it writeBarItems() calls are not valid
func newI3barProtocolSenderSendHeaders() i3barProtocolSender {
	sender := i3barProtocolSender{}
	sender.sendHeaders()
	return sender
}

// send "headers" to i3bar. without these, the output we write in doRefresh() won't be valid
func (i i3barProtocolSender) sendHeaders() {
	// add also empty items line, so we don't have to special case the first payload line we send
	// (first payload line has to be "[...]", second ",[...]")
	fmt.Println(`{ "version": 1, "click_events": true }` + "\n[\n[]")
}

func (i i3barProtocolSender) writeBarItems(items []barItem) {
	itemsJSON, err := json.Marshal(items)
	if err != nil {
		panic(err)
	}

	fmt.Println("," + string(itemsJSON))
}
