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

		if verbose {
			fmt.Println("args")
			fmt.Println(args)
		}

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

		retrieveMetadata := make(map[string]map[string]string)

		allTypes := []string{
			"apps",
			"authproviders",
			"datasources",
			"pages",
			"profiles",
			"themes",
		}

		camelCasings := make(map[string]string)

		// Put in default camel casings into the map
		for _, typeName := range allTypes {
			camelCasings[typeName] = typeName
		}
		// Add in camel-casing exceptions
		camelCasings.authproviders = "authProviders"
		camelCasings.datasources = "dataSources"

		fetchAll := false

		fetchAllByType := make(map[string]bool)
		fetchByType := make(map[string]string)

		// If we have args, build up our list of args to retrieve
		if len(args) > 0 {

			for _, path := range args {
				// Handle global wildcard
				if strings.Contains(path, "**/*") {
					fetchAll = true
					break
				} else if strings.Contains(path, "}/*") && strings.Contains(path, "{") {
					// Handle multi-type wildcard, e.g. '{pages,themes}/*'
					startTypes, endTypes := strings.Index(path, "{"), strings.Index(path, "}")
					desiredTypes := path[startTypes:endTypes]
					for typeName := strings.Split(desiredTypes, ",") {
						fetchAllByType[typeName] = true
					}
				}
				else if strings.Contains(path, "/*") && !strings.Contains(path, "{") {
					// Handle single-type wildcard, e.g. 'pages/*'
					wildcardIndex := strings.Index(path, "/*")
					// If the wildcard is the only /, then everything up to the slash
					// is the name of the type to retrieve
					if strings.Index(path, "/") == wildcardIndex {
						typeName := path[0:wildcardIndex]
						fetchAllByType[typeName] = true
					} 
					// If we have other / chars, split on the last index prior to the wildcard
					else {
						if stringParts := strings.Split(path, "/"); len(stringParts) > 1 {
							typeName := stringParts[len(stringParts) - 2]
							fetchAllByType[typeName] = true
						}
					}
				}
				// // Handle individual file requests
				// else {

				// }
			}
		} 

		if fetchAll {
			// fetch all of each type,
			// which is accomplished by sending an empty map	
			for _, metadataType := range allTypes {
				retrieveMetadata[camelCasings[metadataType]] = make(map[string]string)
			}
		} 
		// Process individual and/or type-wildcard requests
		else {
			for _, metadataType := range allTypes {
				if fetchAllByType[metadataType] == true {
					// If we have a "fetch all" for this type,
					// ignore individual type requests and fetch it all
					retrieveMetadata[camelCasings[metadataType]] = make(map[string]string)
				} else if fetchByType[metadataType] != nil && len(fetchByType[metadataType]) > 0 {
					// Otherwise, process individual file requests
					retrieveMetadata[camelCasings[metadataType]] = fetchByType[metadataType]
				}
			}	
		}
		
		retrieveRequest := &types.RetrieveRequest{}
		retrieveRequest.Metadata = retrieveMetadata

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
			// and then make dirs for all paths up to that point, i.e. pages, apps
			intermediatePath := filepath.Join(targetLocation, strings.Join(pathParts[:],","))
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
			continue
		}

		if verbose {
			fmt.Println("Creating file: " + file.Name)
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
