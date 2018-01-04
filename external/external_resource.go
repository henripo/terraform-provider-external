package external

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/hashicorp/terraform/helper/schema"
)

type Input struct {
	Id           string                 `json:"id,omitempty"`
	Arguments    map[string]interface{} `json:"arguments"`
	OldArguments map[string]interface{} `json:"old_arguments,omitempty"`
}

type Response struct {
	Id        string                 `json:"id"`
	Arguments map[string]interface{} `json:"arguments"`
	Result    map[string]interface{} `json:"result"`
}

func externalResource() *schema.Resource {
	return &schema.Resource{
		Create: resourceCreate,
		Read:   resourceRead,
		Update: resourceUpdate,
		Delete: resourceDelete,

		Schema: map[string]*schema.Schema{
			"program": &schema.Schema{
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"arguments": &schema.Schema{
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"result": &schema.Schema{
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceRead(d *schema.ResourceData, meta interface{}) error {

	programI := d.Get("program").([]interface{})
	arguments := d.Get("arguments").(map[string]interface{})
	input := &Input{
		Id:        d.Id(),
		Arguments: arguments,
	}
	response, err := commandExec(programI, input, "read")
	if err != nil {
		return err
	}

	d.Set("result", response.Result)
	d.Set("arguments", response.Arguments)
	d.SetId(response.Id)
	return nil
}

func resourceCreate(d *schema.ResourceData, meta interface{}) error {
	programI := d.Get("program").([]interface{})
	arguments := d.Get("arguments").(map[string]interface{})

	input := &Input{
		Arguments: arguments,
	}

	response, err := commandExec(programI, input, "create")
	if err != nil {
		return err
	}

	d.Set("result", response.Result)
	d.Set("arguments", response.Arguments)
	d.SetId(response.Id)
	return nil
}
func resourceUpdate(d *schema.ResourceData, meta interface{}) error {
	programI := d.Get("program").([]interface{})
	o, n := d.GetChange("arguments") //.(map[string]interface{})
	//.(map[string]interface{})
	old_arguments := o.(map[string]interface{})
	arguments := n.(map[string]interface{})
	input := &Input{
		Id:           d.Id(),
		Arguments:    arguments,
		OldArguments: old_arguments,
	}

	response, err := commandExec(programI, input, "update")

	if err != nil {
		return err
	}

	d.Set("result", response.Result)
	d.Set("arguments", response.Arguments)
	d.SetId(response.Id)
	return nil
}
func resourceDelete(d *schema.ResourceData, meta interface{}) error {
	programI := d.Get("program").([]interface{})
	o, n := d.GetChange("arguments")
	old_arguments := o.(map[string]interface{})
	arguments := n.(map[string]interface{})
	input := &Input{
		Id:           d.Id(),
		Arguments:    arguments,
		OldArguments: old_arguments,
	}

	_, err := commandExec(programI, input, "delete")

	if err != nil {
		return err
	}
	return nil
}

func commandExec(programI []interface{}, input *Input, operation string) (*Response, error) {

	if err := validateProgramAttr(programI); err != nil {
		return nil, err
	}

	program := make([]string, len(programI)+1)
	for i, vI := range programI {
		program[i] = vI.(string)
	}
	program[len(programI)] = operation

	cmd := exec.Command(program[0], program[1:]...)

	inputJson, err := json.Marshal(input)

	if err != nil {
		return nil, err
	}

	cmd.Stdin = bytes.NewReader(inputJson)

	resultJson, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.Stderr != nil && len(exitErr.Stderr) > 0 {
				return nil, fmt.Errorf("failed to execute %q: %s", program[0], string(exitErr.Stderr))
			}
			return nil, fmt.Errorf("command %q failed with no error message", program[0])
		} else {
			return nil, fmt.Errorf("failed to execute %q: %s", program[0], err)
		}
	}
	var response Response
	if operation == "delete" {
		return nil, nil
	}
	err = json.Unmarshal(resultJson, &response)
	if err != nil {
		return nil, fmt.Errorf("command %q produced invalid JSON: %s", program[0], err)
	}
	return &response, nil
}
