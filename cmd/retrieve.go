package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"archive/zip"
	"path/filepath"
	"strings"

	"github.com/skuid/skuid/pliny"
	"github.com/skuid/skuid/types"
	"github.com/spf13/cobra"
)

// retrieveCmd represents the retrieve command
var retrieveCmd = &cobra.Command{
	Use:   "retrieve",
	Short: "Retrieve Skuid metadata from a Site into a local directory.",
	Long: "Retrieve Skuid metadata from a Skuid Platform Site and output it into a local directory.",
	Run: func(cmd *cobra.Command, args []string) {

		//login to a Skuid Platform Site
		api, err := pliny.Login(
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

		
		retrieveMetadata := &types.RetrieveMetadata{}
		retrieveMetadata.Apps = make(map[string]string)
		retrieveMetadata.DataSources = make(map[string]string)
		retrieveMetadata.Pages = make(map[string]string)
		retrieveMetadata.Profiles = make(map[string]string)
		retrieveMetadata.Themes = make(map[string]string)

		retrieveRequest := &types.RetrieveRequest{}

		fmt.Println("Retrieving metadata...")

		//query the API for all Skuid metadata of every type
		result, err := api.Connection.Post("/metadata/retrieve", retrieveRequest)

		fmt.Println("Successfully retrieved metadata!")

		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		// unzip the archive into the output directory
		targetDirFriendly := targetDir
		if targetDir == "" {
			targetDirFriendly = "current working directory"
		}
		fmt.Println("Writing results to " + targetDirFriendly + " ...")

		tmpfile, err := ioutil.TempFile("", "skuid")

		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		tmpFileName := tmpfile.Name()
		// write to our new file
		tmpfile.Write(result)

		// unzip the contents of our temp zip file
		err = unzip(tmpFileName, targetDir)
		
		// schedule cleanup of temp file
		defer os.Remove(tmpFileName)

		if err != nil {
			fmt.Println("Error unzipping the retrieved metadata payload")
			fmt.Println(err.Error())
			os.Exit(1)
		}

		fmt.Printf("Results written to %s\n", targetDirFriendly)
	},
}

// Unzips a ZIP archive and recreates the folders and file structure within it locally
func unzip(sourceFileLocation, targetLocation string) error {

	reader, err := zip.OpenReader(sourceFileLocation)

	if err != nil {
		return err
	}

	// If we have a non-empty target directory, ensure it exists
	if targetLocation != "" {
		if err := os.MkdirAll(targetLocation, 0755); err != nil {
			return err
		}
	}

	for _, file := range reader.File {
		path := filepath.Join(targetLocation, file.Name)

		// If this file name contains a /, make sure that we create the
		if pathParts := strings.Split(file.Name, "/"); len(pathParts) > 0 {
			// Remove the actual file name from the slice, 
			// i.e. pages/MyAwesomePage.xml ---> pages
			pathParts = pathParts[:len(pathParts)-1]
			// and then make dirs for all paths up to that point, i.e. pages
			intermediatePaths := filepath.Join(targetLocation, strings.Join(pathParts[:],","))
			if verbose {
				fmt.Println("Creating intermediate directories: " + intermediatePaths + "...")
			}
			os.MkdirAll(intermediatePaths, 0755)
		}

		if file.FileInfo().IsDir() {
			os.MkdirAll(path, file.Mode())
			continue
		}

		if verbose {
			fmt.Println("Creating file: " + file.Name + "...")
		}

		fileReader, err := file.Open()
		if err != nil {
			return err
		}

		defer fileReader.Close()

		targetFile, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)

		if err != nil {
			return err
		}

		defer targetFile.Close()

		if _, err := io.Copy(targetFile, fileReader); err != nil {
			return err
		}
	}

	return nil
}



func init() {
	RootCmd.AddCommand(retrieveCmd)
}
