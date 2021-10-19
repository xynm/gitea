// Copyright 2021 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package models

import (
	"context"
	"fmt"

	"code.gitea.io/gitea/models/db"
	"code.gitea.io/gitea/modules/git"
	"code.gitea.io/gitea/modules/timeutil"
)

// RepoArchiverStatus represents repo archive status
type RepoArchiverStatus int

// enumerate all repo archive statuses
const (
	RepoArchiverGenerating = iota // the archiver is generating
	RepoArchiverReady             // it's ready
)

// RepoArchiver represents all archivers
type RepoArchiver struct {
	ID          int64           `xorm:"pk autoincr"`
	RepoID      int64           `xorm:"index unique(s)"`
	Repo        *Repository     `xorm:"-"`
	Type        git.ArchiveType `xorm:"unique(s)"`
	Status      RepoArchiverStatus
	CommitID    string             `xorm:"VARCHAR(40) unique(s)"`
	CreatedUnix timeutil.TimeStamp `xorm:"INDEX NOT NULL created"`
}

func init() {
	db.RegisterModel(new(RepoArchiver))
}

// LoadRepo loads repository
func (archiver *RepoArchiver) LoadRepo() (*Repository, error) {
	if archiver.Repo != nil {
		return archiver.Repo, nil
	}

	var repo Repository
	has, err := db.GetEngine(db.DefaultContext).ID(archiver.RepoID).Get(&repo)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, ErrRepoNotExist{
			ID: archiver.RepoID,
		}
	}
	return &repo, nil
}

// RelativePath returns relative path
func (archiver *RepoArchiver) RelativePath() (string, error) {
	return fmt.Sprintf("%d/%s/%s.%s", archiver.RepoID, archiver.CommitID[:2], archiver.CommitID, archiver.Type.String()), nil
}

// GetRepoArchiver get an archiver
func GetRepoArchiver(ctx context.Context, repoID int64, tp git.ArchiveType, commitID string) (*RepoArchiver, error) {
	var archiver RepoArchiver
	has, err := db.GetEngine(ctx).Where("repo_id=?", repoID).And("`type`=?", tp).And("commit_id=?", commitID).Get(&archiver)
	if err != nil {
		return nil, err
	}
	if has {
		return &archiver, nil
	}
	return nil, nil
}

// AddRepoArchiver adds an archiver
func AddRepoArchiver(ctx context.Context, archiver *RepoArchiver) error {
	_, err := db.GetEngine(ctx).Insert(archiver)
	return err
}

// UpdateRepoArchiverStatus updates archiver's status
func UpdateRepoArchiverStatus(ctx context.Context, archiver *RepoArchiver) error {
	_, err := db.GetEngine(ctx).ID(archiver.ID).Cols("status").Update(archiver)
	return err
}

// DeleteAllRepoArchives deletes all repo archives records
func DeleteAllRepoArchives() error {
	_, err := db.GetEngine(db.DefaultContext).Where("1=1").Delete(new(RepoArchiver))
	return err
}
