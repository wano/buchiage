package buchiage

import (
	"encoding/json"
	"testing"
)

func Test_Config(t *testing.T) {
	testString := `
{
 "conf_list": [
	{
		"name": "hoge",
		"matcher": "hoge",
		"dest": "parent_hoge",
		"event_name": "Create",
		"bucket": "test_bucket"
	},
	{
		"name": "foobar",
		"matcher": "foobar",
		"dest": "parent_foobar",
		"event_name": "Rename",
		"bucket": "test_bucket"
	}
	]
}
`
	var conf ConfigList
	if err := json.Unmarshal([]byte(testString), &conf); err != nil {
		t.Fatal(err)
	}

	if len(conf.ConfList) != 2 {
		t.Errorf("wont: 2, got: %d", len(conf.ConfList))
	}

}
