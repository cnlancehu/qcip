//go:build windows

package main

import (
	"gopkg.in/toast.v1"
)

func init() {
	notify = func(title string, message string) {
		notification := toast.Notification{
			AppID:   "QCIP",
			Title:   title,
			Message: message,
		}
		err := notification.Push()
		if err != nil {
			errhandle("Error while sending notification: " + err.Error())
		}
	}
}
