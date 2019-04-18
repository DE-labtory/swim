package common_test

import (
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"testing"

	"github.com/DE-labtory/swim/cmd/common"
	"github.com/stretchr/testify/assert"
)

func TestRelativeToAbsolutePath(t *testing.T) {

	testfile1 := "./util.go"
	testabsresult1, err := filepath.Abs(testfile1)
	assert.NoError(t, err)
	testabs1, err := common.RelativeToAbsolutePath(testfile1)

	assert.NoError(t, err)
	assert.Equal(t, testabs1, testabsresult1)

	testfile2 := "../README.md"
	testabsresult2, err := filepath.Abs(testfile2)
	assert.NoError(t, err)

	testabs2, err := common.RelativeToAbsolutePath(testfile2)

	assert.NoError(t, err)
	assert.Equal(t, testabs2, testabsresult2)

	// 남의 홈패스에 뭐가있는지 알길이 없으니 하나 만들었다 지움
	usr, err := user.Current()
	assert.NoError(t, err)

	testfile3 := usr.HomeDir + "/test.txt"

	_, err = os.Stat(usr.HomeDir)
	if os.IsNotExist(err) {
		file, err := os.Create(testfile3)
		assert.NoError(t, err)
		defer file.Close()
	}

	err = ioutil.WriteFile(testfile3, []byte("test"), os.ModePerm)
	assert.NoError(t, err)

	testfile4 := "~/test.txt"

	testabs3, err := common.RelativeToAbsolutePath(testfile4)
	assert.NoError(t, err)
	assert.Equal(t, testfile3, testabs3)

	err = os.Remove(testfile3)
	assert.NoError(t, err)
}

func TestRelativeToAbsolutePath_WhenGivenPathIsAbsolute(t *testing.T) {
	sshPath := "/iAmRoot"

	absPath, err := common.RelativeToAbsolutePath(sshPath)

	assert.NoError(t, err)
	assert.Equal(t, sshPath, absPath)
}

func TestRelativeToAbsolutePath_WhenGivenPathWithOnlyName(t *testing.T) {
	sshPath := "test-dir"

	absPath, err := common.RelativeToAbsolutePath(sshPath)
	currentPath, _ := filepath.Abs(".")

	assert.NoError(t, err)
	assert.Equal(t, path.Join(currentPath, sshPath), absPath)
}

func TestRelativeToAbsolutePath_WhenGivenPathIsEmpty(t *testing.T) {
	sshPath := ""

	absPath, err := common.RelativeToAbsolutePath(sshPath)

	assert.Equal(t, nil, err)
	assert.Equal(t, "", absPath)
}
