package main

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/bashdi/gherkin-backup/backup"
)

var (
	intervall    int    = 120
	targetFolder string = "targetFolder"
	backupFolder string = "backupFolder"
)

func main() {
	getParameters()

	log.Println("gherkin-backup started")
	backupProcess := backup.NewBackupProcess(targetFolder, backupFolder, &backup.BackupAlgorithmGHK1{})

	for {
		err := backupProcess.DoIt()
		if err != nil {
			log.Println(err)
		}
		time.Sleep(time.Duration(intervall) * time.Second)
	}
}

func getParameters() {
	parameters := os.Args[1:]

	if len(parameters) != 3 {
		log.Fatal("Wrong/missing parameters: app \"[targetFolder]\" \"[backupFolder]\" [intervall in seconds]")
	}

	targetFolder = parameters[0]
	backupFolder = parameters[1]
	intervallParameter, err := strconv.Atoi(parameters[2])
	if err != nil {
		log.Fatalf("\"%s\" is not a integer\n", parameters[2])
	}
	intervall = intervallParameter
}
