package main

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_setHost(t *testing.T) {
	assert := assert.New(t)

	root := "http://localhost:5500/addpath"

	u, _ := url.Parse("http://test:5000/test?how=to")
	k, _ := url.Parse("http://localhost:5500/addpath/test?how=to")

	assert.Equal(k.String(),
		setHost(root, u), "they should be equal")
}

func Test_propHeadertoLog(t *testing.T) {
	assert := assert.New(t)
	tar := "dsid=dialog-session-id"
	assert.Equal("\"dsid\": \"${header:dialog-session-id}\", ",
		propHeadertoLog(tar, ""), "they should be equal")

	tar = "dsid=dialog-session-id"
	assert.Equal("\"dsid\": \"${header:dialog-session-id}\", ",
		propHeadertoLog(tar, ""), "they should be equal")
}
