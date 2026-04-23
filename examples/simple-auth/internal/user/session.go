package user

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/franchb/htmgo/framework/v2/h"
	"github.com/franchb/htmgo/framework/v2/service"
	"simpleauth/internal/db"
)

type CreatedSession struct {
	Id         string
	Expiration time.Time
	UserId     int64
}

func CreateSession(ctx *h.RequestContext, userId int64) (CreatedSession, error) {
	sessionId, err := GenerateSessionID()

	if err != nil {
		return CreatedSession{}, err
	}

	// create a session in the database
	queries := service.Get[db.Queries](ctx.ServiceLocator())

	created := CreatedSession{
		Id:         sessionId,
		Expiration: time.Now().Add(time.Hour * 24),
		UserId:     userId,
	}

	err = queries.CreateSession(context.Background(), db.CreateSessionParams{
		UserID:    created.UserId,
		SessionID: created.Id,
		ExpiresAt: created.Expiration.Format(time.RFC3339),
	})

	if err != nil {
		return CreatedSession{}, err
	}

	return created, nil
}

func GetUserFromSession(ctx *h.RequestContext) (db.User, error) {
	sessionId := ctx.Fiber.Cookies("session_id")
	if sessionId == "" {
		return db.User{}, errors.New("no session cookie")
	}
	queries := service.Get[db.Queries](ctx.ServiceLocator())
	user, err := queries.GetUserByToken(context.Background(), sessionId)
	if err != nil {
		return db.User{}, err
	}
	return user, nil
}

func WriteSessionCookie(ctx *h.RequestContext, session CreatedSession) {
	cookie := fiber.Cookie{
		Name:     "session_id",
		Value:    session.Id,
		HTTPOnly: true,
		Secure:   true,
		SameSite: fiber.CookieSameSiteStrictMode,
		Expires:  session.Expiration,
		Path:     "/",
	}
	ctx.SetCookie(&cookie)
}

func GenerateSessionID() (string, error) {
	// Create a byte slice for storing the random bytes
	bytes := make([]byte, 32) // 32 bytes = 256 bits, which is a secure length
	// Read random bytes from crypto/rand
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	// Encode to hexadecimal to get a string representation
	return hex.EncodeToString(bytes), nil
}
