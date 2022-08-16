package backup

import (
	"time"
)

type BackupAlgorithm interface {
	IsFileChanged(targetFile, backupFolder string) (bool, error)
	Backup(targetFiles []FilechangeInfo, backupFolder string, timestamp time.Time) error
	Restore(targetFolder, backupFolder, timestamp time.Time) error
}
