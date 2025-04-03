package middleware

import (
	"fmt"
	"gopkg.in/telebot.v4"
)

func ImageValidator(next telebot.HandlerFunc) telebot.HandlerFunc {
	const maxImageSizeKilobytes = 5000 * 1000

	return func(c telebot.Context) error {
		doc := c.Message().Photo
		if doc != nil && doc.FileSize > maxImageSizeKilobytes {
			_ = c.Reply(fmt.Sprintf("Your image is too large. Maximum allowed size is %d :(", maxImageSizeKilobytes))

			return nil
		}
		return next(c)
	}
}
