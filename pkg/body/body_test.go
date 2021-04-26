package body_test

import (
	"fmt"
	"nydus/pkg/body"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnmarshal(t *testing.T) {
	tem := `{"Test": "Test"}`

	actual, _ := body.Unmarshal([]byte(tem), "application/json")
	expected1 := map[string]interface{}{"Test": "Test"}
	assert.Equal(t, expected1, actual, "기대값과 결과값이 다릅니다.")

	tem = `[{"Test": "Test"}, {"Test2": "Test2"}]`

	actual, _ = body.Unmarshal([]byte(tem), "application/json")
	fmt.Println(actual)
	mar, _ := body.Marshal(actual, "application/json")
	fmt.Println(string(mar))
	expected2 := []interface{}([]interface{}{map[string]interface{}{"Test": "Test"}, map[string]interface{}{"Test2": "Test2"}})
	assert.Equal(t, expected2, actual, "기대값과 결과값이 다릅니다.")

	actual, _ = body.Unmarshal([]byte(""), "")
	fmt.Println(actual)
	expected3 := ""
	assert.Equal(t, expected3, actual, "기대값과 결과값이 다릅니다.")
}
