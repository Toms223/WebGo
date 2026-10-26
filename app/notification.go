package app

import (
	"encoding/json"

	"github.com/Toms223/WebGo/internal/errors"
)

// Notification is shown with NotificationService.New. Path opens when it is
// clicked; the "goapp" key of Data is reserved.
type Notification struct {
	Title string `json:"title"`

	Path string `json:"path"`

	Lang string `json:"lang,omitempty"`

	Badge string `json:"badge,omitempty"`

	Body string `json:"body,omitempty"`

	Tag string `json:"tag,omitempty"`

	Icon string `json:"icon,omitempty"`

	Image string `json:"image,omitempty"`

	Data map[string]any `json:"data"`

	Renotify bool `json:"renotify,omitempty"`

	RequireInteraction bool `json:"requireInteraction,omitempty"`

	Silent bool `json:"silent,omitempty"`

	Vibrate []int `json:"vibrate,omitempty"`

	Actions []NotificationAction `json:"actions,omitempty"`
}

type NotificationAction struct {
	Action string `json:"action"`

	Title string `json:"title"`

	Icon string `json:"icon,omitempty"`

	Path string `json:"path"`
}

type NotificationSubscription struct {
	Endpoint string `json:"endpoint"`

	Keys struct {
		Auth string `json:"auth"`

		P256dh string `json:"p256dh"`
	} `json:"keys"`
}

type NotificationPermission string

const (
	NotificationDefault NotificationPermission = "default"

	NotificationGranted NotificationPermission = "granted"

	NotificationDenied NotificationPermission = "denied"

	NotificationNotSupported NotificationPermission = "unsupported"
)

// NotificationService shows notifications and manages push subscriptions in the
// browser.
type NotificationService struct{}

func (s NotificationService) Permission() NotificationPermission {
	notification := Window().Get("Notification")
	if !notification.Truthy() {
		return NotificationNotSupported
	}

	return NotificationPermission(notification.Get("permission").String())
}

func (s NotificationService) RequestPermission() NotificationPermission {
	notification := Window().Get("Notification")
	if !notification.Truthy() {
		return NotificationNotSupported
	}

	permission := make(chan string, 1)
	defer close(permission)

	notification.Call("requestPermission").Then(func(v Value) {
		permission <- v.String()
	})

	return NotificationPermission(<-permission)
}

func (s NotificationService) New(n Notification) {
	notification, _ := json.Marshal(n)
	Window().Call("goappNewNotification", string(notification))
}

// Subscribe returns a push subscription for the VAPID public key, blocking
// until the browser answers. It fails when the key is empty or push is
// unsupported.
func (s NotificationService) Subscribe(vapIDPublicKey string) (NotificationSubscription, error) {
	if vapIDPublicKey == "" {
		return NotificationSubscription{}, errors.New("vapid public key is empty")
	}

	subc := make(chan string, 1)
	defer close(subc)

	Window().Call("goappSubscribePushNotifications", vapIDPublicKey).Then(func(v Value) {
		subc <- v.String()
	})

	jsSub := <-subc
	if jsSub == "" {
		return NotificationSubscription{}, errors.
			New("push notifications are not supported by the browser")
	}

	var sub NotificationSubscription
	err := json.Unmarshal([]byte(jsSub), &sub)
	if err != nil {
		return NotificationSubscription{}, errors.
			New("failed to decode push notification subscription").Wrap(err)
	}
	return sub, nil
}
