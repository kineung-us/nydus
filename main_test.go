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
