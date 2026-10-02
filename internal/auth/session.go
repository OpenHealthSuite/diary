package auth

import (
	"crypto/rand"
	"crypto/sha256"

	"github.com/boj/redistore"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"github.com/openhealthsuite/diary/internal/config"
	"github.com/rs/zerolog/log"
)

const redisPoolMaxIdle = 10

type rediStore struct {
	*redistore.RediStore
}

func (s *rediStore) Options(options sessions.Options) {
	s.RediStore.Options = options.ToGorillaOptions()
}

func newSessionMiddleware(cfg *config.ServerConfiguration) (gin.HandlerFunc, error) {
	store, err := newRedisStore(cfg)
	if err != nil {
		return nil, err
	}
	return sessions.Sessions(cfg.SessionCookieName, store), nil
}

func newRedisStore(cfg *config.ServerConfiguration) (redis.Store, error) {
	authKey, encryptKey, err := sessionKeyPairs(cfg.SessionSecret)
	if err != nil {
		return nil, err
	}

	str, err := redistore.NewRediStoreWithURL(redisPoolMaxIdle, cfg.SessionRedisUrl, authKey, encryptKey)
	if err != nil {
		return nil, err
	}

	conn := str.Pool.Get()
	if _, err = conn.Do("PING"); err != nil {
		conn.Close()
		str.Close()
		return nil, err
	}
	conn.Close()

	return &rediStore{RediStore: str}, nil
}

func sessionKeyPairs(secret string) (authKey, encryptKey []byte, err error) {
	material := []byte(secret)
	if secret == "" {
		material = make([]byte, 32)
		if _, err := rand.Read(material); err != nil {
			return nil, nil, err
		}
		log.Warn().Msg("no OPENFOODDIARY_SESSION_SECRET set, generated an ephemeral one. Sessions will not survive a restart and will not be shared between replicas.")
	}
	authSum := sha256.Sum256(append([]byte("ofd-session-auth:"), material...))
	encryptSum := sha256.Sum256(append([]byte("ofd-session-encrypt:"), material...))
	return authSum[:], encryptSum[:], nil
}
