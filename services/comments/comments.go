// Copyright 2019 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package comments

import (
	"code.gitea.io/gitea/models"
	"code.gitea.io/gitea/models/db"
	"code.gitea.io/gitea/models/issues"
	"code.gitea.io/gitea/modules/notification"
	"code.gitea.io/gitea/modules/timeutil"
)

// CreateIssueComment creates a plain issue comment.
func CreateIssueComment(doer *models.User, repo *models.Repository, issue *models.Issue, content string, attachments []string) (*models.Comment, error) {
	comment, err := models.CreateComment(&models.CreateCommentOptions{
		Type:        models.CommentTypeComment,
		Doer:        doer,
		Repo:        repo,
		Issue:       issue,
		Content:     content,
		Attachments: attachments,
	})
	if err != nil {
		return nil, err
	}
	err = issues.SaveIssueContentHistory(db.GetEngine(db.DefaultContext), doer.ID, issue.ID, comment.ID, timeutil.TimeStampNow(), comment.Content, true)
	if err != nil {
		return nil, err
	}

	mentions, err := issue.FindAndUpdateIssueMentions(db.DefaultContext, doer, comment.Content)
	if err != nil {
		return nil, err
	}

	notification.NotifyCreateIssueComment(doer, repo, issue, comment, mentions)

	return comment, nil
}

// UpdateComment updates information of comment.
func UpdateComment(c *models.Comment, doer *models.User, oldContent string) error {
	if err := models.UpdateComment(c, doer); err != nil {
		return err
	}

	if c.Type == models.CommentTypeComment && c.Content != oldContent {
		err := issues.SaveIssueContentHistory(db.GetEngine(db.DefaultContext), doer.ID, c.IssueID, c.ID, timeutil.TimeStampNow(), c.Content, false)
		if err != nil {
			return err
		}
	}

	notification.NotifyUpdateComment(doer, c, oldContent)

	return nil
}

// DeleteComment deletes the comment
func DeleteComment(doer *models.User, comment *models.Comment) error {
	if err := models.DeleteComment(comment); err != nil {
		return err
	}

	notification.NotifyDeleteComment(doer, comment)

	return nil
}
