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

	for _, targetFile := range newOrDifferentFiles {
		_, err := bp.backupAlgorithm.Backup(bp.targetFolder, targetFile, bp.backupFolder, timestamp)
		if err != nil {
			log.Println(err)
		}
		log.Printf("Backup created for %s", targetFile)
	}

	return nil
}

func (bp *BackupProcess) checkForFilechanges(targetFolder string) ([]string, error) {
	changedFilesList := []string{}
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

		isNewOrDifferent, err := bp.backupAlgorithm.IsFileNewOrDifferent(fullPath, bp.backupFolder)
		if err != nil {
			log.Println(err)
			continue
		}

		if isNewOrDifferent {
			changedFilesList = append(changedFilesList, fullPath)
		}
	}

	return changedFilesList, nil
}
