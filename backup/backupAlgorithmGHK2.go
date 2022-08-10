package backup

import (
	"encoding/hex"
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

func (ghk2 *BackupAlgorithmGHK2) IsFileNewOrDifferent(targetFile, backupFolder string) (bool, error) {
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

func (ghk2 *BackupAlgorithmGHK2) Backup(targetFolder, targetFile, backupFolder string, timestamp time.Time) (bool, error) {
	db, err := badger.Open(ghk2.badgerOptions)
	if err != nil {
		return false, err
	}
	defer db.Close()

	filedepotPath := filepath.Join(backupFolder, filedepotFolderName)
	os.Mkdir(filedepotPath, os.ModePerm)
	if err != nil {
		return false, err
	}

	hash, err := utilities.GetMd5HashForFile(targetFile)
	if err != nil {
		return false, err
	}

	hashString := hex.EncodeToString(hash[:])

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
		extension := filepath.Ext(targetFile)
		depotFile := filepath.Join(filedepotPath, hashString)
		depotFile = depotFile + extension
		utilities.CopyFile(targetFile, depotFile)

		err := db.Update(func(txn *badger.Txn) error {
			return txn.Set([]byte(hashString), []byte(depotFile))
		})
		if err != nil {
			return false, err
		}
	}

	return true, nil
}
