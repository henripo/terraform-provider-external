package external

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

const testResourceConfig_basic = `
resource "external" "test" {
  program = ["%s"]

  arguments = {
	value = "pizza"
  }
}

output "argument_value" {
  value = "${external.test.arguments.value}"
}

output "id" {
  value = "${external.test.id}"
}
output "result" {
	value = "${external.test.result.name}"
}
`

func TestResource_basic(t *testing.T) {
	programPath, err := buildResourceTestProgram()
	if err != nil {
		t.Fatal(err)
		return
	}

	resource.UnitTest(t, resource.TestCase{
		Providers: testProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: fmt.Sprintf(testResourceConfig_basic, programPath),
				Check: func(s *terraform.State) error {
					_, ok := s.RootModule().Resources["external.test"]
					if !ok {
						return fmt.Errorf("missing resource")
					}

					outputs := s.RootModule().Outputs

					if outputs["argument_value"] == nil {
						return fmt.Errorf("missing 'argument_value' output")
					}
					if outputs["id"] == nil {
						return fmt.Errorf("missing 'id' output")
					}
					if outputs["result"] == nil {
						return fmt.Errorf("missing 'result' output")
					}

					if outputs["argument_value"].Value != "pizza" {
						return fmt.Errorf(
							"'argument' output is %q; want 'pizza'",
							outputs["argument_value"].Value,
						)
					}
					if outputs["id"].Value != "mock" {
						return fmt.Errorf(
							"'id' output is %q; want 'mock'",
							outputs["id"].Value,
						)
					}
					if outputs["result"].Value != "mock" {
						return fmt.Errorf(
							"'result' output is %q; want 'mock'",
							outputs["result"].Value,
						)
					}

					return nil
				},
			},
		},
	})
}

/*
const testDataSourceConfig_error = `
data "external" "test" {
  program = ["%s"]

  query = {
    fail = "true"
  }
}
`

func TestDataSource_error(t *testing.T) {
	programPath, err := buildDataSourceTestProgram()
	if err != nil {
		t.Fatal(err)
		return
	}

	resource.UnitTest(t, resource.TestCase{
		Providers: testProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config:      fmt.Sprintf(testDataSourceConfig_error, programPath),
				ExpectError: regexp.MustCompile("I was asked to fail"),
			},
		},
	})
}
*/
func buildResourceTestProgram() (string, error) {
	cmd := exec.Command(
		"go", "install",
		"github.com/terraform-providers/terraform-provider-external/external/test-programs/tf-acc-external-resource",
	)
	err := cmd.Run()

	if err != nil {
		return "", fmt.Errorf("failed to build test stub program: %s", err)
	}

	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = filepath.Join(os.Getenv("HOME") + "/go")
	}

	programPath := path.Join(
		filepath.SplitList(gopath)[0], "bin", "tf-acc-external-resource",
	)
	return programPath, nil
}
