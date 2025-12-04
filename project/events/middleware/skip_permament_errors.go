package middleware

import (
	"errors"

	"github.com/ThreeDotsLabs/watermill/message"
)
type PermanentError interface {
	IsPermanent() bool
}

func SkipPermanentErrorsMiddleware(h message.HandlerFunc) message.HandlerFunc {
	return func(msg *message.Message) ([]*message.Message, error) {
		msgs, err := h(msg)

		var permErr PermanentError
		if errors.As(err, &permErr) && permErr.IsPermanent() {
			return nil, nil
		}

		return msgs, err
	}
}