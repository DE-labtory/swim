package conf_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/DE-labtory/swim"

	"github.com/DE-labtory/swim/conf"
	"github.com/stretchr/testify/assert"
)

func TestGetConfiguration(t *testing.T) {
	path := os.Getenv("GOPATH") + "/src/github.com/DE-labtory/swim/conf"
	confFileName := "/config-test.yaml"
	defer os.Remove(path + confFileName)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, os.ModePerm)
	}

	// please leave this whitespace as space not tab
	ConfJson := []byte(`
  swimconfig:
    maxlocalcount: 2
    maxnsacounter: 8
    t: 3000
    acktimeout: 3000
    k: 2
    bindaddress: 127.0.0.1
    bindport: 50000
  suspicionconfig:
    k: 2
    min: 3000
    max: 10000
  messageendpointconfig:
    encryptionenabled: false
    sendtimeout: 5000
    callbackcollectinterval: 7000
  member:
    id:
      id: "hello_id"
    `)

	err := ioutil.WriteFile(path+confFileName, ConfJson, os.ModePerm)
	assert.NoError(t, err)

	conf.SetConfigPath(path + confFileName)
	config := conf.GetConfiguration()
	assert.Equal(t, config.SWIMConfig.MaxlocalCount, 2)
	assert.Equal(t, config.SWIMConfig.MaxNsaCounter, 8)
	assert.Equal(t, config.SWIMConfig.T, 3000)
	assert.Equal(t, config.SWIMConfig.AckTimeOut, 3000)
	assert.Equal(t, config.SWIMConfig.K, 2)
	assert.Equal(t, config.SWIMConfig.BindAddress, "127.0.0.1")
	assert.Equal(t, config.SWIMConfig.BindPort, 50000)

	assert.Equal(t, config.SuspicionConfig.K, 2)
	assert.Equal(t, config.SuspicionConfig.Min.Nanoseconds(), int64(3000))
	assert.Equal(t, config.SuspicionConfig.Max.Nanoseconds(), int64(10000))

	assert.Equal(t, config.MessageEndpointConfig.EncryptionEnabled, false)
	assert.Equal(t, config.MessageEndpointConfig.SendTimeout.Nanoseconds(), int64(5000))
	assert.Equal(t, config.MessageEndpointConfig.CallbackCollectInterval.Nanoseconds(), int64(7000))

	assert.Equal(t, config.Member.ID, swim.MemberID{ID: "hello_id"})
}
