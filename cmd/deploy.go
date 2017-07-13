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

		// If target directory is not provided,
		// use the current directory's contents
		targetDirFriendly := targetDir
		if targetDir == "" {
			targetDir = "."
			targetDirFriendly, err = filepath.Abs(filepath.Dir(os.Args[0]))
			if err != nil {
				log.Fatal(err)
			}
		}
		if verbose {
			fmt.Println("Deploying site from", targetDirFriendly)
		}

		// Create a temporary directory into which to put our ZIP
		tmpDir, err := ioutil.TempDir("", "skuid-deploy")
		if err != nil {
		    panic(err)
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
		    panic(err)
		}

		reader, err := os.Open(outFilePath)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Deploying metadata...")

		_, err = api.Connection.MakeRequest(http.MethodPost, "/metadata/deploy", reader, "application/zip")

		if err != nil {
			fmt.Println("Error deploying metadata:", err.Error())
			os.Exit(1)
		}

		fmt.Println("Successfully deployed metadata to Skuid Site.")
	},
}

func init() {
	RootCmd.AddCommand(deployCmd)
}
