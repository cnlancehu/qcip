//go:build windows

package main

import (
	_ "embed"
	"os"
	"time"

	"gopkg.in/toast.v1"
)

var (
	//go:embed static/success.ico
	successIcon []byte
	//go:embed static/failed.ico
	failedIcon []byte
	err        error
)

func init() {
	// 系统为 windows 时，启用 winnotify 并显示对应的帮助信息
	var errOccurred = false
	notifyHelpMsg = "\n  -n  --winnotify\tSend notifacation cards, only available on Windows"
	notifyErrCheck := func() {
		if err != nil {
			err = nil
			errOccurred = true
		}
	}
	notify = func(title string, message string, succeed bool) {
		genNotifyAndLog := func(succeed bool, notification toast.Notification, iconPath string) toast.Notification {
			notification = toast.Notification{
				AppID:   "QCIP",
				Title:   title,
				Message: "Some error occurred, please check the error log for details",
			}
			if iconPath != "" {
				notification.Icon = iconPath
			}
			if !succeed {
				err = os.WriteFile("qciperrlog.txt", []byte("QCIP error log | "+time.Now().Format("2006-01-02 15:04:05")+"\n"+message), 0644)
				if err != nil {
					notification.Message = message
					return notification
				}
				errLogPath, _ := os.Getwd()
				errLogPath += "\\qciperrlog.txt"
				notification.Actions = []toast.Action{
					{
						Type:      "protocol",
						Label:     "Show error logs",
						Arguments: errLogPath,
					},
				}
			} else {
				notification.Message = message
			}
			return notification
		}
		var iconFile *os.File
		iconFile, err = os.CreateTemp("", "qcip-*.ico")
		notifyErrCheck()
		if succeed {
			_, err = iconFile.Write(successIcon)
			notifyErrCheck()
		} else {
			_, err = iconFile.Write(failedIcon)
			notifyErrCheck()
		}
		err = iconFile.Close()
		notifyErrCheck()
		iconPath := iconFile.Name()
		var notification toast.Notification
		if !errOccurred {
			notification = genNotifyAndLog(succeed, notification, iconPath)
		} else {
			notification = genNotifyAndLog(succeed, notification, "")
		}
		err = notification.Push()
		notifyErrCheck()
		if errOccurred {
			errOutput("Error occurred when sending notification cards")
		}
	}
}
