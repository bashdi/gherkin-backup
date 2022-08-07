package backup

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/bashdi/gherkin-backup/utilities"
)

type BackupAlgorithmGHK1 struct {
	lastChangeMap map[string]time.Time
	currentBackup time.Time
}

func (ghk1 *BackupAlgorithmGHK1) IsFileNewOrDifferent(targetFile string) (bool, error) {
	//lastChangeMap init, if nil
	if ghk1.lastChangeMap == nil {
		ghk1.lastChangeMap = make(map[string]time.Time)
	}

	fileinfo, err := os.Stat(targetFile)
	if err != nil {
		return false, err
	}

	currentModDate := fileinfo.ModTime()
	if currentModDate != ghk1.lastChangeMap[targetFile] {
		ghk1.lastChangeMap[targetFile] = currentModDate
		return true, nil
	}
	return false, nil
}

func (ghk1 *BackupAlgorithmGHK1) Backup(targetFolder, targetFile, backupFolder string, timestamp time.Time) (bool, error) {
	if ghk1.currentBackup == timestamp {
		return true, nil
	}

	backupFolder = filepath.Join(backupFolder, timestamp.Format("2006_01_02_15_04_05"))
	os.Mkdir(backupFolder, os.ModePerm)
	ghk1.backup(targetFolder, backupFolder)

	return true, nil
}

func (ghk1 *BackupAlgorithmGHK1) backup(targetFolder, backupFolder string) (bool, error) {
	files, err := os.ReadDir(targetFolder)
	if err != nil {
		log.Println(err)
		return false, err
	}

	for _, file := range files {
		if file.IsDir() {
			childBackupFolder := filepath.Join(backupFolder, file.Name())
			os.Mkdir(childBackupFolder, os.ModePerm)
			ghk1.backup(filepath.Join(targetFolder, file.Name()), childBackupFolder)
			continue
		}

		targetFile := filepath.Join(targetFolder, file.Name())
		backupFile := filepath.Join(backupFolder, file.Name())
		utilities.CopyFile(targetFile, backupFile)
	}

	return true, nil
}
