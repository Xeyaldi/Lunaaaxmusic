/*
 * TgMusicBot - Telegram Music Bot
 *  Copyright (c) 2025-2026 Ashok Shau
 *
 *  Licensed under GNU GPL v3
 *  See https://github.com/AshokShau/TgMusicBot
 */

package handlers

import (
	"ashokshau/tgmusic/src/utils"
	"ashokshau/tgmusic/src/vc"
	"fmt"
	"time"

	"ashokshau/tgmusic/src/core/cache"

	"github.com/AshokShau/gotdbot"
)

const reloadCooldown = 3 * time.Minute

var reloadRateLimit = cache.NewCache[time.Time](reloadCooldown)

func reloadAdminCacheHandler(c *gotdbot.Client, ctx *gotdbot.Context) error {
	m := ctx.EffectiveMessage
	if m.IsPrivate() {
		return gotdbot.EndGroups
	}

	reloadKey := fmt.Sprintf("reload:%d", m.ChatId)
	if lastUsed, ok := reloadRateLimit.Get(reloadKey); ok {
		timePassed := time.Since(lastUsed)
		if timePassed < reloadCooldown {
			remaining := int((reloadCooldown - timePassed).Seconds())
			// Gözləmə müddəti mesajı
			_, _ = m.ReplyText(c, fmt.Sprintf("⏳ <b>Gözləyin:</b> Bu əmrdən yenidən istifadə etmək üçün lütfən <code>%s</code> gözləyin.", utils.SecToMin(remaining)), nil)
			return nil
		}
	}

	reloadRateLimit.Set(reloadKey, time.Now())

	// Yenilənmə başlayanda çıxan mesaj
	reply, err := m.ReplyText(c, "🔄 <i>Administrator keşi yenilənir, lütfən gözləyin...</i>", nil)
	if err != nil {
		c.Logger.Warn("Failed to send reloading message for chat", "chat_id", m.ChatId, "error", err)
		return gotdbot.EndGroups
	}

	cache.ClearAdminCache(m.ChatId)
	vc.Calls.UpdateInviteLink(m.ChatId, "")

	admins, err := cache.GetAdmins(c, m.ChatId, true)
	if err != nil {
		c.Logger.Warn("Failed to reload the admin cache for chat", "chat_id", m.ChatId, "error", err)
		// Xəta baş verərsə mesaj dəyişir
		_, _ = reply.EditText(c, "💥 <b>Xəta:</b> Administrator keşini yeniləmək mümkün olmadı.", nil)
		return gotdbot.EndGroups
	}

	c.Logger.Info("Reloaded admins for chat", "count", len(admins), "chat_id", m.ChatId)
	// Uğurla başa çatanda mesaj dəyişir
	_, _ = reply.EditText(c, "✅ <b>Uğurlu:</b> Administrator keşi uğurla yeniləndi! ⚡", nil)
	return gotdbot.EndGroups
}

// privacyHandler handles the /privacy command.
func privacyHandler(c *gotdbot.Client, ctx *gotdbot.Context) error {
	m := ctx.EffectiveMessage
	botName := c.Me.FirstName

	// Məxfilik siyasəti mətni (Gözəl dizayn edilmiş forması)
	text := fmt.Sprintf("🛡 <b>%s üçün Məxfilik Siyasəti</b>\n\n"+
		"📋 <b>1. Data Saxlanılması:</b>\nCihazınızda heç bir şəxsi məlumat saxlanılmır. Sizin axtarış və ya qrup fəaliyyətiniz izlənilmir.\n\n"+
		"🔍 <b>2. Məlumatların Toplanması:</b>\nBiz yalnız musiqi xidmətlərini təmin etmək üçün sizin Telegram <code>User ID</code> və <code>Chat ID</code> məlumatlarınızı oxuyuruq. Ad, soyad, telefon nömrəsi və ya məkan qeydə alınmır.\n\n"+
		"⚙️ <b>3. İstifadə:</b>\nToplanan məlumatlar sırf botun funksiyalarının düzgün işləməsi üçündür. Reklam və ya kommersiya məqsədilə istifadə edilə bilməz.\n\n"+
		"🤝 <b>4. Paylaşım:</b>\nMəlumatlarınız üçüncü şəxslərlə paylaşılmır, satılmır və ya ötürülmür.\n\n"+
		"🔒 <b>5. Təhlükəsizlik:</b>\nMəlumatları qorumaq üçün standart şifrələmədən istifadə olunur. Lakin unutmayın ki, internetdə heç bir servis 100%% təhlükəsiz deyil.\n\n"+
		"🍪 <b>6. Kukilər (Cookies):</b>\n%s kukilərdən və ya digər izləmə texnologiyalarından istifadə etmir.\n\n"+
		"🌐 <b>7. Üçüncü Tərəflər:</b>\nTelegram-ın özü istisna olmaqla, heç bir kənar məlumat toplayıcı sistemlə inteqrasiyamız yoxdur.\n\n"+
		"💡 <b>8. Sizin Hüquqlarınız:</b>\nİstədiyiniz vaxt botu bloklayaraq giriş icrasını ləğv edə və ya məlumatların silinməsini tələb edə bilərsiniz.\n\n"+
		"📢 <b>9. Yenilənmələr:</b>\nSiyasətdə hər hansı bir dəyişiklik olarsa, bot vasitəsilə elan ediləcək.\n\n"+
		"💬 <b>10. Əlaqə:</b>\nSualınız var? Bizim <a href=\"https://t.me/GuardxSupport\">Dəstək Qrupumuza</a> müraciət edin.\n\n"+
		"──────────────────\n"+
		"⚠️ <i><b>Qeyd:</b> Bu siyasət %s ilə təhlükəsiz və hörmətli bir təcrübə yaşamağınızı təmin edir.</i>", botName, botName, botName)

	_, err := m.ReplyText(c, text, &gotdbot.SendTextMessageOpts{ParseMode: "html", DisableWebPagePreview: true})
	return err
}
