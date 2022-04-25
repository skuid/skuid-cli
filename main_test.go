package main_test

import (
	"log"
	"os"
	"os/exec"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/joho/godotenv"

	tides "github.com/skuid/tides"
)

func setup() {
	// make sure we have git credentials in environment variables
	username := os.Getenv(tides.ENV_GITHUB_USERNAME)
	pac := os.Getenv(tides.ENV_PERSONAL_ACCESS_TOKEN)

	if username == "" {
		log.Fatalf("Github Username required for integration tests. Environment Variable: %v", tides.ENV_GITHUB_USERNAME)
	}

	if pac == "" {
		log.Fatalf("Github Personal Access Token required for integration tests. Environment Variable: %v", tides.ENV_PERSONAL_ACCESS_TOKEN)
	}

	// clone monorepo
	_, err := git.PlainClone("/tmp/tides/tests", false, &git.CloneOptions{
		URL:      "https://github.com/skuid/skuid-monorepo",
		Progress: os.Stdout,
		Auth: &http.BasicAuth{
			Username: username,
			Password: pac,
		},
	})

	if err != nil {
		log.Fatal(err)
	}

	// let's go back when we're done
	currDir, _ := os.Getwd()
	defer os.Chdir(currDir)

	err = os.Chdir("/tmp/tides/tests/skuid-monorepo/skuid-core")
	if err != nil {
		log.Fatalf("Unable to chdir into skuid-core: %v", err.Error())
	}

	err = exec.Command("make", "start").Run()
	if err != nil {
		log.Fatalf("Unable to run skuid-core make start: %v", err.Error())
	}

	// pool, err := dockertest.NewPool("TestGetRetrievePlan")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// resource, err := pool.RunWithOptions(&dockertest.RunOptions{
	// 	Repository: "postgres",
	// 	Tag:        "11",
	// 	Env: []string{
	// 		"POSTGRES_USER=test",
	// 		"POSTGRES_PASSWORD=test",
	// 		"listen_addresses = '*'",
	// 	},
	// }, func(config *docker.HostConfig) {
	// 	// set AutoRemove to true so that stopped container goes away by itself
	// 	config.AutoRemove = true
	// 	config.RestartPolicy = docker.RestartPolicy{
	// 		Name: "no",
	// 	}
	// })

}

func cleanup() {
	err := os.RemoveAll("/tmp/tides/tests")
	if err != nil {
		log.Fatal(err)
	}
}

func loadEnv() {
	err := godotenv.Load(".testenv")
	if err != nil {
		log.Fatal(err)
	}
}

func TestMain(m *testing.M) {
	loadEnv()
	setup()
	defer cleanup()

	m.Run()

}
