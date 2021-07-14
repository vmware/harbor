package oidc

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/oidc/dao"
)

// FixEmptySubIss remediates the issue https://github.com/goharbor/harbor/issues/15241
// by restoring the subiss via the persisted token
func FixEmptySubIss(ctx context.Context) error {
	metaMgr := NewMetaMgr()
	meta, err := metaMgr.GetBySubIss(ctx, "", "")
	if errors.IsNotFoundErr(err) {
		log.Info("Not found any records with empty subiss, good to go.")
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to query for OIDC info with empty subiss: %v", err)
	}
	log.Infof("Found record with empty subiss, user ID: %d, trying to restore...", meta.UserID)
	key, err := keyLoader.encryptKey()
	if err != nil {
		return fmt.Errorf("failed to load the key for encryption/decryption： %v", err)
	}
	tokenStr, err := utils.ReversibleDecrypt(meta.Token, key)
	if err != nil {
		return fmt.Errorf("failed to decrypt token： %v", err)
	}
	// Extract the subject issuer from persisted ID token
	tok := &Token{}
	err = json.Unmarshal(([]byte)(tokenStr), tok)
	if err != nil {
		return fmt.Errorf("failed to decode token: %v", err)
	}
	rawIDToken := tok.RawIDToken
	parts := strings.Split(rawIDToken, ".")
	if len(parts) < 2 {
		return fmt.Errorf("malformed jwt, got %d parts", len(parts))
	}
	data, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return fmt.Errorf("malformed jwt: %v", err)
	}
	p := &struct {
		Issuer  string `json:"iss"`
		Subject string `json:"sub"`
	}{}
	err = json.Unmarshal(data, p)
	if err != nil {
		return fmt.Errorf("failed to extract subject and issuer from ID token: %v", err)
	}
	newRec := &models.OIDCUser{
		ID:     meta.ID,
		SubIss: p.Subject + p.Issuer,
	}
	metaDao := dao.NewMetaDao()
	if err := metaDao.Update(ctx, newRec, "subiss"); err != nil {
		return fmt.Errorf("failed to update meta info in DB: %v", err)
	}
	log.Infof("Restored subiss for user, id: %d", meta.UserID)
	return nil
}
