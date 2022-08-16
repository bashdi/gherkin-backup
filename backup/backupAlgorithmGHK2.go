package backup

import (
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/bashdi/gherkin-backup/utilities"
	"github.com/dgraph-io/badger/v3"
)

const (
	filedepotFolderName string = "filedepot"
)

type BackupAlgorithmGHK2 struct {
	isBadgerOptionsInit bool
	badgerOptions       badger.Options
}

func (ghk2 *BackupAlgorithmGHK2) IsFileChanged(targetFile, backupFolder string) (bool, error) {
	if !ghk2.isBadgerOptionsInit {
		ghk2.isBadgerOptionsInit = true

		dbDir := filepath.Join(backupFolder, "ghk2")
		ghk2.badgerOptions = badger.DefaultOptions(dbDir)
		ghk2.badgerOptions.MetricsEnabled = false
	}

	db, err := badger.Open(ghk2.badgerOptions)
	if err != nil {
		return false, err
	}
	defer db.Close()

	isNewOrModified := true

	badgerErr := db.Update(func(txn *badger.Txn) error {
		file, err := os.Stat(targetFile)
		if err != nil {
			return err
		}

		item, err := txn.Get([]byte(targetFile))
		if err != nil {
			//Creating new entry
			err := txn.Set([]byte(targetFile), []byte(strconv.FormatInt(file.ModTime().Unix(), 10)))
			if err != nil {
				return err
			}
			log.Printf("New file: %s\n", targetFile)
			return nil
		}

		//Compare modifyDate from file and db entry
		var modifyDateEntry int64
		item.Value(func(val []byte) error {
			modifyDateEntry, _ = strconv.ParseInt(string(val), 10, 64)
			return nil
		})

		if modifyDateEntry != file.ModTime().Unix() {
			log.Printf("File change: %s\n", targetFile)

			err := txn.Set([]byte(targetFile), []byte(strconv.FormatInt(file.ModTime().Unix(), 10)))
			if err != nil {
				return err
			}
			return nil
		}
		isNewOrModified = false
		return nil
	})
	if badgerErr != nil {
		return false, err
	}

	return isNewOrModified, nil
}

func (ghk2 *BackupAlgorithmGHK2) Backup(targetFiles []FilechangeInfo, backupFolder string, timestamp time.Time) error {
	db, err := badger.Open(ghk2.badgerOptions)
	if err != nil {
		return err
	}
	defer db.Close()

	backupFiles := []Ghk2File{}
	isAnyFileChange := false

	for _, targetFile := range targetFiles {
		filedepotPath := filepath.Join(backupFolder, filedepotFolderName)
		os.Mkdir(filedepotPath, os.ModePerm)
		if err != nil {
			return err
		}

		hash, err := utilities.GetMd5HashForFile(targetFile.Path)
		if err != nil {
			return err
		}

		hashString := hex.EncodeToString(hash[:])

		backupFiles = append(backupFiles, Ghk2File{Path: targetFile.Path, Hash: hashString})
		if !targetFile.IsChanged {
			continue
		}
		isAnyFileChange = true

		//check if file allready in depot or needs to copied
		isFileInDepot := true
		db.View(func(txn *badger.Txn) error {
			_, err := txn.Get([]byte(hashString))
			if err != nil {
				isFileInDepot = false
			}
			return nil
		})

		if !isFileInDepot {
			extension := filepath.Ext(targetFile.Path)
			depotFile := filepath.Join(filedepotPath, hashString)
			depotFile = depotFile + extension
			utilities.CopyFile(targetFile.Path, depotFile)

			err := db.Update(func(txn *badger.Txn) error {
				return txn.Set([]byte(hashString), []byte(depotFile))
			})
			if err != nil {
				return err
			}
		}
	}

	if isAnyFileChange {
		backupState := Ghk2BackupState{Timestamp: timestamp.Unix(), Files: backupFiles}

		file, err := json.MarshalIndent(backupState, "", "")
		if err != nil {
			return nil
		}

		err = ioutil.WriteFile(filepath.Join(backupFolder, timestamp.Format("2006_01_02_15_04_05")+".json"), file, 0644)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ghk2 *BackupAlgorithmGHK2) Restore(targetFolder, backupFolder, timestamp time.Time) error {
	return nil
}

type Ghk2BackupState struct {
	Timestamp int64
	Files     []Ghk2File
}

type Ghk2File struct {
	Path string
	Hash string
}
