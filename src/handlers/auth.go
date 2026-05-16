/*
 * TgMusicBot - Telegram Music Bot
 *  Copyright (c) 2025-2026 Ashok Shau
 *
 *  Licensed under GNU GPL v3
 *  See https://github.com/AshokShau/TgMusicBot
 */

package handlers

import (
	"ashokshau/tgmusic/src/core/cache"
	"ashokshau/tgmusic/src/core/db"
	"fmt"

	td "github.com/AshokShau/gotdbot"
)

func authListHandler(c *td.Client, ctx *td.Context) error {
	if !adminMode(c, ctx) {
		return td.EndGroups
	}

	m := ctx.EffectiveMessage
	if m.IsPrivate() {
		return nil
	}

	chatID := m.ChatId

	authUser := db.Instance.GetAuthUsers(chatID)
	if authUser == nil || len(authUser) == 0 {
		_, _ = m.ReplyText(c, "❌ <i>Səlahiyyətli istifadəçi tapılmadı!</i>", nil)
		return nil
	}

	text := "🛡 <b><u>SƏLAHİYYƏTLİ İSTİFADƏÇİLƏR</u></b>\n\n"
	for _, uid := range authUser {
		text += fmt.Sprintf("👤 • <a href=\"tg://user?id=%d\">%d</a>\n", uid, uid)
	}

	_, _ = m.ReplyText(c, text, replyOpts)
	return td.EndGroups
}

func addAuthHandler(c *td.Client, ctx *td.Context) error {
	if !adminMode(c, ctx) {
		return td.EndGroups
	}

	m := ctx.EffectiveMessage
	if m.IsPrivate() {
		return td.EndGroups
	}

	chatID := m.ChatId

	UserStatus, err := cache.GetUserAdmin(c, chatID, m.SenderID(), false)
	if err != nil {
		c.Logger.Warn("GetUserAdmin error", "error", err)
		_, _ = m.ReplyText(c, "⚠️ <i>Adminlik statusu yoxlanıla bilmədi!</i>", nil)
		return td.EndGroups
	}

	switch UserStatus.Status.(type) {
	case *td.ChatMemberStatusCreator, *td.ChatMemberStatusAdministrator:
	default:
		_, _ = m.ReplyText(c, "🚫 <b>Xəta:</b> Bu əmrdən istifadə etmək üçün <u>Administrator</u> olmalısınız!", nil)
		return td.EndGroups
	}

	userID, err := getTargetUserID(c, m)
	if err != nil {
		_, _ = m.ReplyText(c, err.Error(), nil)
		return nil
	}

	if db.Instance.IsAuthUser(chatID, userID) {
		_, _ = m.ReplyText(c, "ℹ️ Bu istifadəçiyə <b>artıq</b> səlahiyyət verilib.", nil)
		return nil
	}

	if err = db.Instance.AddAuthUser(chatID, userID); err != nil {
		c.Logger.Error("Failed to add authorized user", "error", err)
		_, _ = m.ReplyText(c, "💥 <b>Uğursuz cəhd:</b> İstifadəçiyə səlahiyyət verilərkən xəta yarandı.", nil)
		return nil
	}

	_, err = m.ReplyText(c, fmt.Sprintf("✅ <b>Uğurlu:</b> <code>%d</code> ID-li istifadəçiyə səlahiyyət verildi! ✨", userID), nil)
	return err
}

func removeAuthHandler(c *td.Client, ctx *td.Context) error {
	if !adminMode(c, ctx) {
		return td.EndGroups
	}

	m := ctx.EffectiveMessage
	if m.IsPrivate() {
		return td.EndGroups
	}

	chatID := m.ChatId

	UserStatus, err := cache.GetUserAdmin(c, chatID, m.SenderID(), false)
	if err != nil {
		c.Logger.Warn("GetUserAdmin error", "error", err)
		_, _ = m.ReplyText(c, "⚠️ <i>Adminlik statusu yoxlanıla bilmədi!</i>", nil)
		return td.EndGroups
	}

	switch UserStatus.Status.(type) {
	case *td.ChatMemberStatusCreator, *td.ChatMemberStatusAdministrator:
	default:
		_, _ = m.ReplyText(c, "🚫 <b>Xəta:</b> Bu əmrdən istifadə etmək üçün <u>Administrator</u> olmalısınız!", nil)
		return td.EndGroups
	}

	userID, err := getTargetUserID(c, m)
	if err != nil {
		_, _ = m.ReplyText(c, err.Error(), nil)
		return nil
	}

	if !db.Instance.IsAuthUser(chatID, userID) {
		_, _ = m.ReplyText(c, "ℹ️ Bu istifadəçi onsuz da səlahiyyətli siyahısında <b>deyil</b>.", nil)
		return nil
	}

	if err := db.Instance.RemoveAuthUser(chatID, userID); err != nil {
		c.Logger.Error("Failed to remove authorized user", "error", err)
		_, _ = m.ReplyText(c, "💥 <b>Uğursuz cəhd:</b> Səlahiyyət ləğv edilərkən xəta yarandı.", nil)
		return nil
	}

	_, err = m.ReplyText(c, fmt.Sprintf("🗑 <b>Məlumat:</b> <code>%d</code> ID-li istifadəçinin səlahiyyəti ləğv edildi! ❌", userID), nil)
	return err
}
