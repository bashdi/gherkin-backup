package backup

import (
	"log"
	"os"
	"path/filepath"
	"time"
)

type BackupProcess struct {
	targetFolder    string
	backupFolder    string
	backupAlgorithm BackupAlgorithm
}

func NewBackupProcess(targetFolder string, backupFolder string, backupAlgorithm BackupAlgorithm) *BackupProcess {
	return &BackupProcess{targetFolder: targetFolder, backupFolder: backupFolder, backupAlgorithm: backupAlgorithm}
}

func (bp *BackupProcess) DoIt() error {

	newOrDifferentFiles, err := bp.checkForFilechanges(bp.targetFolder)
	if err != nil {
		return err
	}

	timestamp := time.Now()

	err = bp.backupAlgorithm.Backup(newOrDifferentFiles, bp.backupFolder, timestamp)
	return err
}

func (bp *BackupProcess) checkForFilechanges(targetFolder string) ([]FilechangeInfo, error) {
	changedFilesList := []FilechangeInfo{}
	files, err := os.ReadDir(targetFolder)
	if err != nil {
		return changedFilesList, err
	}

	for _, file := range files {
		fullPath := filepath.Join(targetFolder, file.Name())
		if file.IsDir() {
			childDirChangedFilesList, _ := bp.checkForFilechanges(fullPath)
			changedFilesList = append(changedFilesList, childDirChangedFilesList...)
			continue
		}

		isNewOrDifferent, err := bp.backupAlgorithm.IsFileChanged(fullPath, bp.backupFolder)
		if err != nil {
			log.Println(err)
			continue
		}

		if isNewOrDifferent {
			log.Printf("File: %s changed or new", fullPath)
		}

		changedFilesList = append(changedFilesList, FilechangeInfo{Path: fullPath, IsChanged: isNewOrDifferent})
	}

	return changedFilesList, nil
}
