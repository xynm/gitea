// Copyright 2021 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package auth

import (
	"strings"

	"code.gitea.io/gitea/models"
	"code.gitea.io/gitea/models/db"
	"code.gitea.io/gitea/models/login"
	"code.gitea.io/gitea/modules/log"

	// Register the sources
	_ "code.gitea.io/gitea/services/auth/source/db"
	_ "code.gitea.io/gitea/services/auth/source/ldap"
	_ "code.gitea.io/gitea/services/auth/source/oauth2"
	_ "code.gitea.io/gitea/services/auth/source/pam"
	_ "code.gitea.io/gitea/services/auth/source/smtp"
	_ "code.gitea.io/gitea/services/auth/source/sspi"
)

// UserSignIn validates user name and password.
func UserSignIn(username, password string) (*models.User, *login.Source, error) {
	var user *models.User
	if strings.Contains(username, "@") {
		user = &models.User{Email: strings.ToLower(strings.TrimSpace(username))}
		// check same email
		cnt, err := db.Count(user)
		if err != nil {
			return nil, nil, err
		}
		if cnt > 1 {
			return nil, nil, models.ErrEmailAlreadyUsed{
				Email: user.Email,
			}
		}
	} else {
		trimmedUsername := strings.TrimSpace(username)
		if len(trimmedUsername) == 0 {
			return nil, nil, models.ErrUserNotExist{Name: username}
		}

		user = &models.User{LowerName: strings.ToLower(trimmedUsername)}
	}

	hasUser, err := models.GetUser(user)
	if err != nil {
		return nil, nil, err
	}

	if hasUser {
		source, err := login.GetSourceByID(user.LoginSource)
		if err != nil {
			return nil, nil, err
		}

		if !source.IsActive {
			return nil, nil, models.ErrLoginSourceNotActived
		}

		authenticator, ok := source.Cfg.(PasswordAuthenticator)
		if !ok {
			return nil, nil, models.ErrUnsupportedLoginType
		}

		user, err := authenticator.Authenticate(user, username, password)
		if err != nil {
			return nil, nil, err
		}

		// WARN: DON'T check user.IsActive, that will be checked on reqSign so that
		// user could be hint to resend confirm email.
		if user.ProhibitLogin {
			return nil, nil, models.ErrUserProhibitLogin{UID: user.ID, Name: user.Name}
		}

		return user, source, nil
	}

	sources, err := login.AllActiveSources()
	if err != nil {
		return nil, nil, err
	}

	for _, source := range sources {
		if !source.IsActive {
			// don't try to authenticate non-active sources
			continue
		}

		authenticator, ok := source.Cfg.(PasswordAuthenticator)
		if !ok {
			continue
		}

		authUser, err := authenticator.Authenticate(nil, username, password)

		if err == nil {
			if !authUser.ProhibitLogin {
				return authUser, source, nil
			}
			err = models.ErrUserProhibitLogin{UID: authUser.ID, Name: authUser.Name}
		}

		if models.IsErrUserNotExist(err) {
			log.Debug("Failed to login '%s' via '%s': %v", username, source.Name, err)
		} else {
			log.Warn("Failed to login '%s' via '%s': %v", username, source.Name, err)
		}
	}

	return nil, nil, models.ErrUserNotExist{Name: username}
}
