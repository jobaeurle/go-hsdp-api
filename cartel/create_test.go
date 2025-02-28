package cartel

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreate(t *testing.T) {
	var createResponse = `{
  "message": [
    {
      "eip_address": null,
      "instance_id": "i-xxfbdf005781fa900",
      "ip_address": "192.168.2.106",
      "name": "cadence.dev",
      "role": "container-host"
    }
  ],
  "result": "Success"
}`

	teardown, err := setup(t, &Config{
		Token:  sharedToken,
		Secret: sharedSecret,
		Host:   "foo",
		NoTLS:  true,
	})

	muxCartel.HandleFunc("/v3/api/create", endpointMocker([]byte(sharedSecret),
		createResponse))

	defer teardown()

	if err != nil {
		t.Fatal(err)
	}

	sr, resp, err := client.Create("foo.dev",
		VolumeEncryption(true),
		VolumesAndSize(1, 50),
		SecurityGroups("foo", "bar"),
		UserGroups("andy"))
	if !assert.NotNil(t, resp) {
		return
	}
	if !assert.NotNil(t, sr) {
		return
	}
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, true, sr.Success())
	assert.Equal(t, "i-xxfbdf005781fa900", sr.InstanceID())
	assert.Equal(t, "192.168.2.106", sr.IPAddress())
}

func TestAlreadyExist(t *testing.T) {
	var createResponse = `{
  "code": 400,
  "description": "Host named foo.dev already exists!"
}`
	teardown, err := setup(t, &Config{
		Token:  sharedToken,
		Secret: sharedSecret,
		Host:   "foo",
		NoTLS:  true,
	})

	muxCartel.HandleFunc("/v3/api/create", endpointMocker([]byte(sharedSecret),
		createResponse,
		http.StatusBadRequest))

	defer teardown()

	if err != nil {
		t.Fatal(err)
	}

	_, resp, err := client.Create("foo.dev",
		VolumeEncryption(true),
		VolumesAndSize(1, 50),
		SecurityGroups("foo", "bar"),
		UserGroups("andy"))
	if !assert.NotNil(t, resp) {
		return
	}
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Equal(t, ErrHostnameAlreadyExists, err)
}
