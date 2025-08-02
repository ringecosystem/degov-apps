package services

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/patrickmn/go-cache"
	"github.com/ringecosystem/degov-apps/graph/model"
	"github.com/ringecosystem/degov-apps/internal/config"
	"github.com/ringecosystem/degov-apps/internal/database"
	"github.com/spruceid/siwe-go"
	"gorm.io/gorm"
)

type AuthService struct {
	db         *gorm.DB
	nonceCache *cache.Cache
}

func NewAuthService() *AuthService {
	c := cache.New(3*time.Minute, 5*time.Minute)

	return &AuthService{
		db:         database.GetDB(),
		nonceCache: c,
	}
}

func (s *AuthService) Nonce(input model.GetNonceInput) (string, error) {
	var length int
	if input.Length == nil || *input.Length < 6 {
		length = 32
	} else {
		length = int(*input.Length)
	}
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	nonce := hex.EncodeToString(bytes)

	// put nonce in cache, expires in 3 minutes
	s.nonceCache.Set(nonce, true, 3*time.Minute)
	return nonce, nil
}

func (s *AuthService) Login(input model.LoginInput) (model.LoginOutput, error) {
	message, err := siwe.ParseMessage(input.Message)
	if err != nil {
		err = fmt.Errorf("parse message err: %v", err)
		return model.LoginOutput{}, err
	}
	verify, err := message.ValidNow()
	if err != nil {
		err = fmt.Errorf("message valid failed: %v", err)
		return model.LoginOutput{}, err
	}

	if !verify {
		err = fmt.Errorf("verify message fail")
		return model.LoginOutput{}, err
	}

	//# must open
	nonce := message.GetNonce()
	slog.Debug("login nonce", "nonce", nonce)
	enableCheckNonce := true
	cfg := config.GetConfig()
	if cfg.GetAppEnv().IsDevelopment() {
		enableCheckNonce = cfg.GetStringWithDefault("UNSAFE_ENABLE_VERIFY_NONCE_ON_LOGIN", "true") == "true"
	}
	if enableCheckNonce {
		// check nonce in cache
		_, found := s.nonceCache.Get(nonce)
		if !found {
			err = fmt.Errorf("invalid or expired nonce")
			return model.LoginOutput{}, err
		}
		// After nonce verification, delete it from cache (one-time use)
		s.nonceCache.Delete(nonce)
	}

	// Create JWT token with claims
	claims := jwt.MapClaims{
		"address": message.GetAddress().Hex(),
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// read secret key from environment variable
	secretKey := []byte(config.GetStringRequired("JWT_SECRET"))
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		err = fmt.Errorf("generate token failed: %v", err)
		return model.LoginOutput{}, err
	}

	return model.LoginOutput{
		Token: tokenString,
	}, nil
}
