package test

import (
	"testing"
	"github.com/gruntwork-io/terratest"
	"fmt"
	"github.com/gruntwork-io/terratest/packer"
	"log"
	"time"
	"net/http"
	"net/url"
	"io/ioutil"
	"strings"
)

// The port numbers used by docker-compose.yml in the couchbase-ami example
var testWebConsolePorts = map[string]int{
	"ubuntu": 8091,
}
var testSyncGatewayPorts = map[string]int{
	"ubuntu": 4984,
}

// The username and password we use in all the examples, mocks, and tests
const usernameForTest = "admin"
const passwordForTest = "password"

func createBaseRandomResourceCollection(t * testing.T) *terratest.RandomResourceCollection {
	resourceCollectionOptions := terratest.NewRandomResourceCollectionOptions()

	randomResourceCollection, err := terratest.CreateRandomResourceCollection(resourceCollectionOptions)
	if err != nil {
		t.Fatalf("Failed to create Random Resource Collection: %s", err.Error())
	}

	return randomResourceCollection
}

func createBaseTerratestOptions(t *testing.T, testName string, folder string, resourceCollection *terratest.RandomResourceCollection) *terratest.TerratestOptions {
	terratestOptions := terratest.NewTerratestOptions()

	terratestOptions.UniqueId = resourceCollection.UniqueId
	terratestOptions.TemplatePath = folder
	terratestOptions.TestName = testName

	return terratestOptions
}

func buildCouchbaseWithPacker(t *testing.T, logger *log.Logger, builderName string, awsRegion string, folderPath string) string {
	templatePath := fmt.Sprintf("%s/couchbase.json", folderPath)

	options := packer.PackerOptions{
		Template: templatePath,
		Only: builderName,
		Vars: map[string]string{
			"aws_region": awsRegion,
		},
	}

	artifactId, err := packer.BuildAmi(options, logger)
	if err != nil {
		t.Fatalf("Failed to build Packer template %s: %v", templatePath, err)
	}

	return artifactId
}

func deploy(t *testing.T, terratestOptions *terratest.TerratestOptions) {
	_, err := terratest.Apply(terratestOptions)
	if err != nil {
		t.Fatalf("Failed to apply templates: %s", err.Error())
	}
}

func HttpPostForm(t *testing.T, postUrl string, postParams url.Values, logger *log.Logger) (int, string, error) {
	logger.Printf("Making an HTTP POST call to URL %s with body %v", postUrl, postParams)

	client := http.Client{
		// By default, Go does not impose a timeout, so an HTTP connection attempt can hang for a LONG time.
		Timeout: 10 * time.Second,
	}

	resp, err := client.PostForm(postUrl, postParams)
	if err != nil {
		return -1, "", err
	}

	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return -1, "", err
	}

	return resp.StatusCode, strings.TrimSpace(string(respBody)), nil
}