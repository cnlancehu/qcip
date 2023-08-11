//go:build windows
// +build windows

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
		notification.Push()
	}
}
