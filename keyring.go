package swim

import (
	"os"

	"github.com/DE-labtory/heimdall"
	"github.com/DE-labtory/heimdall/config"
	"github.com/DE-labtory/heimdall/hecdsa"
	"github.com/DE-labtory/iLogger"
)

/*
secLV
128 -> P256
192 -> P384
256 -> P521
*/

// if the key already exists, create a new one, or load the existing key.
func GenerateKey(keyPath string, SecLv int) (heimdall.PriKey, heimdall.PubKey) {
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		pri, pub := GenerateNewKey(keyPath, SecLv)
		return pri, pub
	}
	pri, pub := LoadKeyPair(keyPath)
	return pri, pub
}

func GenerateNewKey(keyPath string, SecLv int) (heimdall.PriKey, heimdall.PubKey) {
	myConFig, err := config.NewSimpleConfig(SecLv)
	if err != nil {
		iLogger.Panic(nil, err.Error())
	}

	keyGenOpt := myConFig.KeyGenOpt
	pri, err := hecdsa.GenerateKey(keyGenOpt)
	if err != nil {
		iLogger.Panic(nil, err.Error())
	}

	err = hecdsa.StorePriKeyWithoutPwd(pri, keyPath)
	if err != nil {
		iLogger.Panic(nil, err.Error())
	}

	return pri, pri.PublicKey()
}

func LoadKeyPair(keyPath string) (heimdall.PriKey, heimdall.PubKey) {
	pri, err := hecdsa.LoadPriKeyWithoutPwd(keyPath)
	if err != nil {
		iLogger.Panic(nil, err.Error())
	}

	return pri, pri.PublicKey()
}

func GetNodeID(keyPath string) string {
	pri, _ := LoadKeyPair(keyPath)

	return pri.ID()
}
