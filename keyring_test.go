package swim

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeyring_generateKey_notExist(t *testing.T) {
	keyPath := "./.key"
	pri, pub := GenerateKey(keyPath, 256)

	defer os.RemoveAll(keyPath)

	assert.NotNil(t, pri)
	assert.NotNil(t, pub)
}

func TestKeyring_generateKey_exist(t *testing.T) {
	keyPath := "./.key"
	pri1, pub1 := GenerateKey(keyPath, 256)
	pri2, pub2 := GenerateKey(keyPath, 256)

	defer os.RemoveAll(keyPath)

	assert.Equal(t, pri1, pri2)
	assert.Equal(t, pub1, pub2)
}

func TestKeyring_getNodeID(t *testing.T) {
	keyPath := "./.key"
	GenerateKey(keyPath, 256)

	defer os.RemoveAll(keyPath)

	assert.NotNil(t, GetNodeID(keyPath))

}
