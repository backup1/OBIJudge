package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	rice "github.com/GeertJohan/go.rice"
	"github.com/nicksnyder/go-i18n/i18n"
)

const (
	appName         = "obijudge"
	appVersion      = "0.1"
	appInfo         = "Created by Gabriel Simões (simoes.sgabriel@gmail.com)"
	appHelp         = "Usage: %s run OR builddb OR info\nAppend -h or --help to display general or subcommand usage\n"
	appErrorMessage = "[OBIJUDGE] "
)

func main() {
	runCommand := flag.NewFlagSet("run", flag.ExitOnError)
	builddbCommand := flag.NewFlagSet("builddb", flag.ExitOnError)

	portPtr := runCommand.Int("port", 8080, "Port where interface will listen (localhost-only")
	databasePtr := runCommand.String("database", "contests.zip", "Contests database file")
	referencePtr := runCommand.String("reference", "reference.zip", "File where language reference is stored")
	workersPtr := runCommand.Int("workers", 2, "Number of simultaneous judge workers")

	sourcePtr := builddbCommand.String("source", "contests", "Folder where contests data is located")
	targetPtr := builddbCommand.String("target", "contests.zip", "File where the database will be created (erases if already exists)")
	passwordPtr := builddbCommand.String("password", "", "16 letters password to encrypt database (will generate one if empty)")

	if len(os.Args) < 2 {
		fmt.Printf(appHelp, os.Args[0])
		os.Exit(0)
	}

	switch os.Args[1] {
	case "run":
		runCommand.Parse(os.Args[2:])
	case "builddb":
		builddbCommand.Parse(os.Args[2:])
	case "info":
		fmt.Println(appName, "version", appVersion)
		fmt.Println(appInfo)
	case "-h":
		fmt.Printf(appHelp, os.Args[0])
	case "--help":
		fmt.Printf(appHelp, os.Args[0])
	default:
		fmt.Printf(appHelp, os.Args[0])
	}

	logger := log.New(os.Stderr, appErrorMessage, log.Ltime)

	if runCommand.Parsed() {
		err := func() error {
			// setup translations
			localesBox := rice.MustFindBox("locales")
			if err := localesBox.Walk("", func(path string, info os.FileInfo, _ error) error {
				if path == "" {
					return nil
				}

				localeBytes, err := localesBox.Bytes(path)
				if err != nil {
					return err
				}

				return i18n.ParseTranslationFileBytes(path, localeByts)
			}); err != nil {
				return err
			}

			db, err := OpenDatabase(*databasePtr)
			if err != nil {
				return err
			}
			defer db.Close()
			db.Logger = logger

			ref, err := OpenReference(*referencePtr)
			if err != nil {
				return err
			}

			judge := &Judge{NumWorkers: *workersPtr, DB: db}
			judge.Start()
			defer judge.Stop()

			server := &Server{
				Port:      *portPtr,
				DB:        db,
				Reference: ref,
				Judge:     judge,
				Logger:    logger,
			}
			if err := server.Start(); err != nil {
				return err
			}
			defer server.Stop()

			stopChan := make(chan os.Signal, 1)
			signal.Notify(stopChan, os.Interrupt)
			select {
			case <-stopChan:
			}
			return nil
		}()
		if err != nil {
			logger.Print(err)
		}
	}

	if builddbCommand.Parsed() {
		err := BuildDatabase(*sourcePtr, *targetPtr, []byte(*passwordPtr))
		if err != nil {
			logger.Fatal(err)
		}
	}
}
