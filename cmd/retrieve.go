package cmd

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"archive/zip"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"path/filepath"

	"github.com/skuid/skuid/platform"
	"github.com/skuid/skuid/types"
	"github.com/spf13/cobra"
)

var packageFile string

// retrieveCmd represents the retrieve command
var retrieveCmd = &cobra.Command{
	Use:   "retrieve",
	Short: "Retrieve Skuid metadata from a Site into a local directory.",
	Long:  "Retrieve Skuid metadata from a Skuid Platform Site and output it into a local directory.",
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

		retrieveMetadata := make(map[string]map[string]string)
		retrieveRequest := &types.RetrieveRequest{}
		var jsonBytes []byte
		if packageFile == "" {
			// To fetch all metadata of a given type,
			// add an empty map for each type's camel-cased name
			for _, camelCaseName := range types.MetadataTypes {
				retrieveMetadata[camelCaseName] = make(map[string]string)
			}
			retrieveRequest.Metadata = retrieveMetadata
			b, err := json.Marshal(retrieveRequest)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
			jsonBytes = b
		} else {
			content, err := ioutil.ReadFile(packageFile)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
			jsonBytes = []byte(fmt.Sprintf(`{"metadata": %s}`, string(content)))
		}

		fmt.Println("Retrieving metadata...")
		//query the API for all Skuid metadata of every type
		result, err := api.Connection.MakeRequest(
			http.MethodPost,
			"/metadata/retrieve",
			bytes.NewReader(jsonBytes),
			"application/json",
		)

		if err != nil {
			fmt.Println("Error retrieving metadata: ", err.Error())
			os.Exit(1)
		}

		fmt.Println("Successfully retrieved metadata!")

		// unzip the archive into the output directory
		targetDirFriendly := targetDir
		if targetDir == "" {
			targetDirFriendly, err = filepath.Abs(filepath.Dir(os.Args[0]))
			if err != nil {
				log.Fatal(err)
			}
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
		readFileFromZipAndWriteToFilesystem(file, targetLocation)
	}

	return nil
}

func readFileFromZipAndWriteToFilesystem(file *zip.File, targetLocation string) error {
	path := filepath.Join(targetLocation, file.Name)

	// If this file name contains a /, make sure that we create the
	if pathParts := strings.Split(file.Name, "/"); len(pathParts) > 0 {
		// Remove the actual file name from the slice,
		// i.e. pages/MyAwesomePage.xml ---> pages
		pathParts = pathParts[:len(pathParts)-1]
		// and then make dirs for all paths up to that point, i.e. pages, apps
		intermediatePath := filepath.Join(targetLocation, strings.Join(pathParts[:], ","))
		//if the desired directory isn't there, create it
		if _, err := os.Stat(intermediatePath); err != nil {
			if verbose {
				fmt.Println("Creating intermediate directory: " + intermediatePath)
			}
			os.MkdirAll(intermediatePath, 0755)
		}
	}

	if file.FileInfo().IsDir() {
		os.MkdirAll(path, file.Mode())
		return nil
	}

	if verbose {
		fmt.Println("Creating file: " + file.Name)
	}

	fileReader, err := file.Open()
	if err != nil {
		return err
	}

	targetFile, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)

	if err != nil {
		return err
	}

	if _, err := io.Copy(targetFile, fileReader); err != nil {
		targetFile.Close()
		fileReader.Close()
		return err
	}

	if err := targetFile.Close(); err != nil {
		return err
	}

	if err := fileReader.Close(); err != nil {
		return err
	}

	return nil
}

func init() {
	retrieveCmd.Flags().StringVarP(&packageFile, "package", "f", "", "Filename of the package definition.")
	RootCmd.AddCommand(retrieveCmd)
}
