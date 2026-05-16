/*
 * TgMusicBot - Telegram Music Bot
 *  Copyright (c) 2025-2026 Ashok Shau
 *
 *  Licensed under GNU GPL v3
 *  See https://github.com/AshokShau/TgMusicBot
 */

package handlers

import (
	"ashokshau/tgmusic/src/core/db"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	td "github.com/AshokShau/gotdbot"
)

var (
	broadcastCancelFlag atomic.Bool
	broadcastInProgress atomic.Bool
)

func getFloodWait(err error) int {
	if err == nil {
		return 0
	}

	type retryError interface {
		GetRetryAfter() int
	}

	if re, ok := err.(retryError); ok {
		return re.GetRetryAfter()
	}

	if tdErr, ok := err.(*td.Error); ok {
		return tdErr.GetRetryAfter()
	}

	if tdErr, ok := err.(td.Error); ok {
		return tdErr.GetRetryAfter()
	}

	return 0
}

func cancelBroadcastHandler(c *td.Client, ctx *td.Context) error {
	if !isDev(ctx) {
		return td.EndGroups
	}
	m := ctx.EffectiveMessage
	if !broadcastInProgress.Load() {
		_, _ = m.ReplyText(c, "❌ __Hal-hazırda aktiv heç bir yayım yoxdur...__", nil)
		return td.EndGroups
	}

	broadcastCancelFlag.Store(true)
	_, _ = m.ReplyText(c, "🛑 **Yayım uğurla dayandırıldı!**", nil)
	return td.EndGroups
}

func broadcastHandler(c *td.Client, ctx *td.Context) error {
	if !isDev(ctx) {
		return td.EndGroups
	}

	m := ctx.EffectiveMessage
	if broadcastInProgress.Load() {
		_, _ = m.ReplyText(c, "⚠️ `Bir yayım artıq davam etməkdədir.` zəhmət olmasa gözləyin...", nil)
		return td.EndGroups
	}

	reply, err := m.GetRepliedMessage(c)
	if err != nil {
		usage := `📣 𝘠𝘢𝘺ı𝘮 𝘦𝘵𝘮ə𝘬 üçü𝘯 𝘻ə𝘩𝘮ə𝒕 𝘰𝘭𝘮𝘢𝘴𝘢 𝘣𝘪𝘳 𝘮𝘦𝘴𝘢𝘫𝘢 𝘤𝘢𝘷𝘢𝘣 (𝘳𝘦𝘱𝘭𝘺) 𝘺𝘢𝘻ı𝘯.

🛠 **İstifadə Qaydası:**
-chat  : 𝘲𝘳𝘶𝘱𝘭𝘢𝘳 🧑‍🤝‍🧑
-user  : 𝘪𝘴𝘵𝘪𝘧𝘢𝘥əç𝘪𝘭ə𝘳 👤
-both  : 𝘲𝘳𝘶𝘱𝘭𝘢𝘳 + 𝘪𝘴𝘵𝘪𝘧𝘢𝘥əç𝘪𝘭ə𝘳 (𝘴essiya) 👥
-copy  : 𝘬öçürmə 𝘮𝘰𝘥𝘶 (𝘤𝘰𝘱𝘺) 📋

📝 __Nümunələr:__
/broadcast
/broadcast -chat
/broadcast -user -copy
`

		_, _ = m.ReplyText(c, usage, nil)
		return td.EndGroups
	}

	args := strings.Fields(Args(m))

	copyMode := false
	mode := "both" // default

	for _, a := range args {
		switch a {
		case "-copy":
			copyMode = true
		case "-chat":
			mode = "chat"
		case "-user":
			mode = "user"
		case "-both":
			mode = "both"
		}
	}

	chats, _ := db.Instance.GetAllChats()
	users, _ := db.Instance.GetAllUsers()

	groupsMap := make(map[int64]bool)
	for _, id := range chats {
		groupsMap[id] = true
	}

	var targets []int64

	switch mode {
	case "chat":
		targets = append(targets, chats...)
	case "user":
		targets = append(targets, users...)
	case "both":
		targets = append(targets, chats...)
		targets = append(targets, users...)
	}

	if len(targets) == 0 {
		_, _ = m.ReplyText(c, "🔍 **Heç bir hədəf tapılmadı.**", nil)
		return td.EndGroups
	}

	broadcastCancelFlag.Store(false)
	broadcastInProgress.Store(true)

	sentMsg, _ := m.ReplyText(c, "🚀 **Yayım başladıldı...** __Xahiş olunur gözləyin.__", nil)

	go func() {
		defer broadcastInProgress.Store(false)

		var failedBuilder strings.Builder
		count, ucount := 0, 0

		for _, chatID := range targets {
			if broadcastCancelFlag.Load() {
				_, _ = sentMsg.EditText(
					c,
					fmt.Sprintf("🛑 **𝘠𝘢𝘺ı𝘮 𝘥𝘢𝘺𝘢𝘯𝘥ı𝘳ı𝘭𝘥ı.**\n\n👥 𝘘𝘳𝘶𝘱𝘭𝘢𝘳: %d\n👤 𝘐𝘴𝘵𝘪𝘧𝘢𝘥əç𝘪𝘭ə𝘳: %d", count, ucount),
					nil,
				)
				return
			}

			var errSend error
			if copyMode {
				_, errSend = reply.Copy(c, chatID, &td.SendCopyOpts{
					ReplyMarkup: reply.ReplyMarkup,
				})
			} else {
				_, errSend = reply.Forward(c, chatID, &td.ForwardMessageOpts{})
			}

			if errSend == nil {
				if groupsMap[chatID] {
					count++
				} else {
					ucount++
				}
				time.Sleep(200 * time.Millisecond)
			} else {
				wait := getFloodWait(errSend)
				if wait > 0 {
					time.Sleep(time.Duration(wait+30) * time.Second)
					continue
				}
				failedBuilder.WriteString(fmt.Sprintf("%d - %v\n", chatID, errSend))
			}
		}

		text := fmt.Sprintf("✅ **Yayım başa çatdı!**\n\n👥 `Qruplar:` %d\n👤 `İstifadəçilər:` %d", count, ucount)
		failedStr := failedBuilder.String()

		if failedStr != "" {
			errFile := filepath.Join(
				os.TempDir(),
				fmt.Sprintf("errors_%d.txt", time.Now().UnixNano()),
			)

			if err := os.WriteFile(errFile, []byte(failedStr), 0644); err == nil {
				defer os.Remove(errFile)

				_, errSendDoc := m.ReplyDocument(
					c,
					td.InputFileLocal{Path: errFile},
					&td.SendDocumentOpts{Caption: text},
				)

				if errSendDoc != nil {
					_, _ = sentMsg.EditText(c, text, nil)
				}
			} else {
				_, _ = sentMsg.EditText(c, text, nil)
			}
		} else {
			_, _ = sentMsg.EditText(c, text, nil)
		}
	}()

	return td.EndGroups
}
