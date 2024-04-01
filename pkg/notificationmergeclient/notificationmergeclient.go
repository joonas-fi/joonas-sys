package notificationmergeclient

import (
	"fmt"
	"net"
	"os"

	. "github.com/function61/gokit/builtin"
	"github.com/function61/gokit/encoding/jsonfile"
)

var (
	// root can send notifications, but there must be a helper symlink which identifies the primary
	// interactive user to route the notifications to. TODO: document who sets this
	ClientForPrimaryInteractiveUser = client{"/run/user.primary-interactive/org.freedesktop.Notifications.Notify_merge.sock"}
)

// JSON field names follow the spec as closely as possible
// https://specifications.freedesktop.org/notification-spec/latest/ar01s09.html#commands
type Notification struct {
	AppName string `json:"app_name"`
	Summary string `json:"summary"`
	Body    string `json:"body"`

	Progress *float64 `json:"progress"` // custom extension
}

func ClientForCurrentUser() client {
	return client{SocketPathCurrentUser()}
}

func SocketPathCurrentUser() string {
	return fmt.Sprintf("/run/user/%d/org.freedesktop.Notifications.Notify_merge.sock", os.Geteuid())
}

type client struct {
	socketPath string
}

// send a notification to the merge server.
// the server either merges related messages or forwards them as individual notifications.
func (c client) Notify(notification Notification) error {
	return ErrorWrap("Notify", notify(notification, c.socketPath))
}

func notify(notification Notification, socketPath string) error {
	mergeServer, err := net.Dial("unix", socketPath)
	if err != nil {
		return err
	}

	if err := jsonfile.Marshal(mergeServer, notification); err != nil {
		return err
	}

	return mergeServer.Close()
}

type notifactionFactory struct {
	appName string
}

func (n notifactionFactory) New(summary string, body string) Notification {
	return Notification{
		AppName: n.appName,
		Summary: summary,
		Body:    body,
	}
}

// progress is between 0.0 and 1.0
func (n notifactionFactory) Progress(summary string, progress float64) Notification {
	return Notification{
		AppName:  n.appName,
		Summary:  summary,
		Progress: &progress,
	}
}

func NotificationFactoryForApp(appName string) notifactionFactory {
	return notifactionFactory{appName}
}
