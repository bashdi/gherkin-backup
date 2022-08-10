package backup

import "time"

type BackupAlgorithm interface {
	IsFileNewOrDifferent(targetFile, backupFolder string) (bool, error)
	Backup(targetFolder, targetFile, backupFolder string, timestamp time.Time) (bool, error)
}
