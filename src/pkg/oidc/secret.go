package oidc

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/pkg/oidc/dao"

	"sync"

	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/lib/log"
)

// SecretVerifyError wraps the different errors happened when verifying a secret for OIDC user.  When seeing this error,
// the caller should consider this an authentication error.
type SecretVerifyError struct {
	cause error
}

func (se *SecretVerifyError) Error() string {
	return fmt.Sprintf("failed to verify the secret: %v", se.cause)
}

func verifyError(err error) error {
	return &SecretVerifyError{err}
}

// SecretManager is the interface for store and verify the secret
type SecretManager interface {
	// VerifySecret verifies the secret and the token associated with it, it refreshes the token in the DB if it's
	// refreshed during the verification.
	VerifySecret(ctx context.Context, username string, secret string) (*UserInfo, error)
}

type keyGetter struct {
	sync.RWMutex
	key string
}

func (kg *keyGetter) encryptKey() (string, error) {
	kg.RLock()
	if kg.key == "" {
		kg.RUnlock()
		kg.Lock()
		defer kg.Unlock()
		if kg.key == "" {
			k, err := config.SecretKey()
			if err != nil {
				return "", err
			}
			kg.key = k
		}
	} else {
		defer kg.RUnlock()
	}
	return kg.key, nil
}

var keyLoader = &keyGetter{}

type defaultManager struct {
	metaDao dao.MetaDAO
}

var m SecretManager = &defaultManager{
	metaDao: dao.NewMetaDao(),
}

// VerifySecret verifies the secret and the token associated with it, it refreshes the token in the DB if it's
// refreshed during the verification.  It returns a populated user model based on the ID token associated with the secret.
func (dm *defaultManager) VerifySecret(ctx context.Context, username string, secret string) (*UserInfo, error) {
	log.Debugf("Verifying the secret for user: %s", username)
	oidcUser, err := dm.metaDao.GetByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to get oidc user info, error: %v", err)
	}
	if oidcUser == nil {
		return nil, fmt.Errorf("user is not onboarded as OIDC user, username: %s", username)
	}
	key, err := keyLoader.encryptKey()
	if err != nil {
		return nil, fmt.Errorf("failed to load the key for encryption/decryption： %v", err)
	}
	plainSecret, err := utils.ReversibleDecrypt(oidcUser.Secret, key)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt secret from DB: %v", err)
	}
	if secret != plainSecret {
		return nil, verifyError(fmt.Errorf("secret mismatch, username: %s", username))
	}
	tokenStr, err := utils.ReversibleDecrypt(oidcUser.Token, key)
	if err != nil {
		return nil, verifyError(err)
	}
	token := &Token{}
	err = json.Unmarshal(([]byte)(tokenStr), token)
	if err != nil {
		return nil, verifyError(err)
	}
	if !token.Valid() {
		log.Debug("Refreshing token")
		token, err = refreshToken(ctx, token)
		if err != nil {
			return nil, fmt.Errorf("failed to refresh token, username: %s, error: %v", username, err)
		}
		tb, err := json.Marshal(token)
		if err != nil {
			return nil, fmt.Errorf("failed to encode the refreshed token, error: %v", err)
		}
		encToken, _ := utils.ReversibleEncrypt(string(tb), key)
		oidcUser.Token = encToken
		// only updates the token column of the record
		err = dm.metaDao.Update(ctx, oidcUser, "token")
		if err != nil {
			log.Errorf("Failed to persist token, user id: %d, error: %v", oidcUser.UserID, err)
		}
		log.Debug("Token refreshed and persisted")
	}
	info, err := UserInfoFromToken(ctx, token)
	if err != nil {
		return nil, verifyError(err)
	}
	return info, nil
}

// VerifySecret calls the manager to verify the secret.
func VerifySecret(ctx context.Context, name string, secret string) (*UserInfo, error) {
	return m.VerifySecret(ctx, name, secret)
}
