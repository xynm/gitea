// Copyright 2020 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package stats

import (
	"code.gitea.io/gitea/models"
	"code.gitea.io/gitea/modules/git"
	"code.gitea.io/gitea/modules/log"
)

// DBIndexer implements Indexer interface to use database's like search
type DBIndexer struct {
}

// Index repository status function
func (db *DBIndexer) Index(id int64) error {
	repo, err := models.GetRepositoryByID(id)
	if err != nil {
		return err
	}
	if repo.IsEmpty {
		return nil
	}

	status, err := repo.GetIndexerStatus(models.RepoIndexerTypeStats)
	if err != nil {
		return err
	}

	gitRepo, err := git.OpenRepository(repo.RepoPath())
	if err != nil {
		return err
	}
	defer gitRepo.Close()

	// Get latest commit for default branch
	commitID, err := gitRepo.GetBranchCommitID(repo.DefaultBranch)
	if err != nil {
		if git.IsErrBranchNotExist(err) || git.IsErrNotExist(err) {
			log.Debug("Unable to get commit ID for default branch %s in %s ... skipping this repository", repo.DefaultBranch, repo.RepoPath())
			return nil
		}
		log.Error("Unable to get commit ID for default branch %s in %s. Error: %v", repo.DefaultBranch, repo.RepoPath(), err)
		return err
	}

	// Do not recalculate stats if already calculated for this commit
	if status.CommitSha == commitID {
		return nil
	}

	// Calculate and save language statistics to database
	stats, err := gitRepo.GetLanguageStats(commitID)
	if err != nil {
		log.Error("Unable to get language stats for ID %s for default branch %s in %s. Error: %v", commitID, repo.DefaultBranch, repo.RepoPath(), err)
		return err
	}
	err = repo.UpdateLanguageStats(commitID, stats)
	if err != nil {
		log.Error("Unable to update language stats for ID %s for default branch %s in %s. Error: %v", commitID, repo.DefaultBranch, repo.RepoPath(), err)
		return err
	}

	log.Debug("DBIndexer completed language stats for ID %s for default branch %s in %s. stats count: %d", commitID, repo.DefaultBranch, repo.RepoPath(), len(stats))
	return nil
}

// Close dummy function
func (db *DBIndexer) Close() {
}
