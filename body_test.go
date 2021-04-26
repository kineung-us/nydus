package test

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnmarshal(t *testing.T) {
	tem := "[{\"custom\":{\"prompt_id\":\"ChitChat\",\"prompt_version\":\"1.0.0\",\"buttons\":[],\"slot_values\":{\"text\":[\"네, 안녕하세요! 제가 도와드릴 일이 있나요?\"]},\"dialog_state\":{\"active_form\":\"\",\"required_slot\":[],\"ask_slot\":[],\"slot_values\":{},\"fail_count\":0}}}]"

	log.Println("TestA running")
}

func TestFoo(t *testing.T) {
	// todo test code
	expected := 1
	actual := 0
	assert.Equal(t, expected, actual, "기대값과 결과값이 다릅니다.")
}
