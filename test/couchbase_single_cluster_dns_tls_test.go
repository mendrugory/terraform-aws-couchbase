package test

import (
	"github.com/gruntwork-io/terratest/modules/aws"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/gruntwork-io/terratest/modules/test-structure"
	"path/filepath"
	"testing"
)

// This domain name is registered in the Gruntwork Phx DevOps account. It also has ACM certs in all regions.
const domainNameForTest = "gruntwork.in"

// We have multiple hosted zones in the Gruntwork Phx DevOps account with the same domain name. This helps
// filter them down to the real public hosted zone for domainNameForTest.
var domainNameTags = map[string]string{"original": "true"}

func TestIntegrationCouchbaseCommunitySingleClusterDnsTlsUbuntu(t *testing.T) {
	t.Parallel()
	testCouchbaseSingleClusterDnsTls(t, "ubuntu", "community")
}

func testCouchbaseSingleClusterDnsTls(t *testing.T, osName string, edition string) {
	examplesFolder := test_structure.CopyTerraformFolderToTemp(t, "../", "examples")
	couchbaseAmiDir := filepath.Join(examplesFolder, "couchbase-ami")
	couchbaseSingleClusterDnsTlsDir := filepath.Join(examplesFolder, "couchbase-cluster-simple-dns-tls")

	test_structure.RunTestStage(t, "setup_ami", func() {
		awsRegion := getRandomAwsRegion(t)
		uniqueId := random.UniqueId()

		amiId := buildCouchbaseAmi(t, osName, couchbaseAmiDir, edition, awsRegion, uniqueId)

		test_structure.SaveAmiId(t, couchbaseSingleClusterDnsTlsDir, amiId)
		test_structure.SaveString(t, couchbaseSingleClusterDnsTlsDir, savedAwsRegion, awsRegion)
		test_structure.SaveString(t, couchbaseSingleClusterDnsTlsDir, savedUniqueId, uniqueId)
	})

	defer test_structure.RunTestStage(t, "teardown", func() {
		terraformOptions := test_structure.LoadTerraformOptions(t, couchbaseSingleClusterDnsTlsDir)
		terraform.Destroy(t, terraformOptions)

		amiId := test_structure.LoadAmiId(t, couchbaseSingleClusterDnsTlsDir)
		awsRegion := test_structure.LoadString(t, couchbaseSingleClusterDnsTlsDir, savedAwsRegion)
		aws.DeleteAmi(t, awsRegion, amiId)
	})

	defer test_structure.RunTestStage(t, "logs", func() {
		terraformOptions := test_structure.LoadTerraformOptions(t, couchbaseSingleClusterDnsTlsDir)
		awsRegion := test_structure.LoadString(t, couchbaseSingleClusterDnsTlsDir, savedAwsRegion)
		testStageLogs(t, terraformOptions, couchbaseClusterVarName, awsRegion)
	})

	test_structure.RunTestStage(t, "setup_deploy", func() {
		amiId := test_structure.LoadAmiId(t, couchbaseSingleClusterDnsTlsDir)
		awsRegion := test_structure.LoadString(t, couchbaseSingleClusterDnsTlsDir, savedAwsRegion)
		uniqueId := test_structure.LoadString(t, couchbaseSingleClusterDnsTlsDir, savedUniqueId)

		terraformOptions := &terraform.Options{
			TerraformDir: couchbaseSingleClusterDnsTlsDir,
			Vars: map[string]interface{}{
				"ami_id":                amiId,
				"domain_name":           domainNameForTest,
				"domain_name_tags":      domainNameTags,
				couchbaseClusterVarName: formatCouchbaseClusterName("single-cluster", uniqueId),
			},
			EnvVars: map[string]string{
				AWS_DEFAULT_REGION_ENV_VAR: awsRegion,
			},
		}

		terraform.InitAndApply(t, terraformOptions)

		test_structure.SaveTerraformOptions(t, couchbaseSingleClusterDnsTlsDir, terraformOptions)
	})

	test_structure.RunTestStage(t, "validation", func() {
		terraformOptions := test_structure.LoadTerraformOptions(t, couchbaseSingleClusterDnsTlsDir)
		validateSingleClusterWorks(t, terraformOptions, couchbaseClusterVarName, "https")
	})
}
