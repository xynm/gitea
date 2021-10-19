// Copyright 2016 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package models

import (
	"code.gitea.io/gitea/models/db"
	"code.gitea.io/gitea/modules/timeutil"
)

// Star represents a starred repo by an user.
type Star struct {
	ID          int64              `xorm:"pk autoincr"`
	UID         int64              `xorm:"UNIQUE(s)"`
	RepoID      int64              `xorm:"UNIQUE(s)"`
	CreatedUnix timeutil.TimeStamp `xorm:"INDEX created"`
}

func init() {
	db.RegisterModel(new(Star))
}

// StarRepo or unstar repository.
func StarRepo(userID, repoID int64, star bool) error {
	sess := db.NewSession(db.DefaultContext)
	defer sess.Close()

	if err := sess.Begin(); err != nil {
		return err
	}

	if star {
		if isStaring(sess, userID, repoID) {
			return nil
		}

		if _, err := sess.Insert(&Star{UID: userID, RepoID: repoID}); err != nil {
			return err
		}
		if _, err := sess.Exec("UPDATE `repository` SET num_stars = num_stars + 1 WHERE id = ?", repoID); err != nil {
			return err
		}
		if _, err := sess.Exec("UPDATE `user` SET num_stars = num_stars + 1 WHERE id = ?", userID); err != nil {
			return err
		}
	} else {
		if !isStaring(sess, userID, repoID) {
			return nil
		}

		if _, err := sess.Delete(&Star{UID: userID, RepoID: repoID}); err != nil {
			return err
		}
		if _, err := sess.Exec("UPDATE `repository` SET num_stars = num_stars - 1 WHERE id = ?", repoID); err != nil {
			return err
		}
		if _, err := sess.Exec("UPDATE `user` SET num_stars = num_stars - 1 WHERE id = ?", userID); err != nil {
			return err
		}
	}

	return sess.Commit()
}

// IsStaring checks if user has starred given repository.
func IsStaring(userID, repoID int64) bool {
	return isStaring(db.GetEngine(db.DefaultContext), userID, repoID)
}

func isStaring(e db.Engine, userID, repoID int64) bool {
	has, _ := e.Get(&Star{UID: userID, RepoID: repoID})
	return has
}

// GetStargazers returns the users that starred the repo.
func (repo *Repository) GetStargazers(opts db.ListOptions) ([]*User, error) {
	sess := db.GetEngine(db.DefaultContext).Where("star.repo_id = ?", repo.ID).
		Join("LEFT", "star", "`user`.id = star.uid")
	if opts.Page > 0 {
		sess = db.SetSessionPagination(sess, &opts)

		users := make([]*User, 0, opts.PageSize)
		return users, sess.Find(&users)
	}

	users := make([]*User, 0, 8)
	return users, sess.Find(&users)
}

// GetStarredRepos returns the repos the user starred.
func (u *User) GetStarredRepos(private bool, page, pageSize int, orderBy string) (repos RepositoryList, err error) {
	if len(orderBy) == 0 {
		orderBy = "updated_unix DESC"
	}
	sess := db.GetEngine(db.DefaultContext).
		Join("INNER", "star", "star.repo_id = repository.id").
		Where("star.uid = ?", u.ID).
		OrderBy(orderBy)

	if !private {
		sess = sess.And("is_private = ?", false)
	}

	if page <= 0 {
		page = 1
	}
	sess.Limit(pageSize, (page-1)*pageSize)

	repos = make([]*Repository, 0, pageSize)

	if err = sess.Find(&repos); err != nil {
		return
	}

	if err = repos.loadAttributes(db.GetEngine(db.DefaultContext)); err != nil {
		return
	}

	return
}

// GetStarredRepoCount returns the numbers of repo the user starred.
func (u *User) GetStarredRepoCount(private bool) (int64, error) {
	sess := db.GetEngine(db.DefaultContext).
		Join("INNER", "star", "star.repo_id = repository.id").
		Where("star.uid = ?", u.ID)

	if !private {
		sess = sess.And("is_private = ?", false)
	}

	return sess.Count(&Repository{})
}
