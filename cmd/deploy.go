package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/pierrre/archivefile/zip"
	"github.com/skuid/skuid/platform"
	"github.com/spf13/cobra"
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy local Skuid metadata to a Skuid Platform Site.",
	Long:  "Deploy Skuid metadata stored within a local file system directory to a Skuid Platform Site.",
	Run: func(cmd *cobra.Command, args []string) {

		api, err := platform.Login(
			host,
			username,
			password,
			apiVersion,
			verbose,
		)

		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		var targetDirFriendly string

		// If target directory is not provided,
		// use the current directory's contents
		if targetDir == "" {
			targetDir = "."
			targetDirFriendly, err = filepath.Abs(filepath.Dir(os.Args[0]))
			if err != nil {
				log.Fatal(err)
			}
		} else {
			targetDirFriendly = targetDir
		}

		if verbose {
			fmt.Println("Deploying site from", targetDirFriendly)
		}

		// Create a temporary directory into which to put our ZIP
		tmpDir, err := ioutil.TempDir("", "skuid-deploy")
		if err != nil {
			log.Fatal("Unable to create a temporary directory for ZIP file")
			log.Fatal(err)
		}
		defer func() {
			_ = os.RemoveAll(tmpDir)
		}()

		outFilePath := filepath.Join(tmpDir, "site.zip")

		progress := func(archivePath string) {
			if verbose {
				fmt.Println("Adding file to ZIP:", archivePath)
			}
		}

		err = zip.ArchiveFile(targetDir, outFilePath, progress)
		if err != nil {
			log.Print("Error creating deployment ZIP archive")
			log.Fatal(err)
		}

		reader, err := os.Open(outFilePath)

		defer reader.Close()

		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Deploying metadata...")

		_, err = api.Connection.MakeRequest(http.MethodPost, "/metadata/deploy", reader, "application/zip")

		if err != nil {
			log.Print("Error deploying metadata")
			log.Fatal(err)
		}

		fmt.Println("Successfully deployed metadata to Skuid Site.")
	},
}

func init() {
	RootCmd.AddCommand(deployCmd)
}
