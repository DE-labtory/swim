package common

import (
	"os/user"
	"path"
	"path/filepath"
	"strings"
)

func RelativeToAbsolutePath(rpath string) (string, error) {
	if rpath == "" {
		return rpath, nil
	}

	absolutePath := ""

	// case ./ ../
	if strings.Contains(rpath, "./") {
		abs, err := filepath.Abs(rpath)
		if err != nil {
			return rpath, err
		}
		return abs, nil
	}

	// case ~/
	if strings.Contains(rpath, "~") {
		i := strings.Index(rpath, "~") // 처음 나온 ~만 반환

		if i > -1 {
			pathRemain := rpath[i+1:]
			usr, err := user.Current()
			if err != nil {
				return rpath, err
			}
			return path.Join(usr.HomeDir, pathRemain), nil

		} else {
			return rpath, nil
		}
	}

	if string(rpath[0]) == "/" {
		return rpath, nil
	}

	if string(rpath[0]) != "." && string(rpath[0]) != "/" {
		currentPath, err := filepath.Abs(".")
		if err != nil {
			return rpath, err
		}

		return path.Join(currentPath, rpath), nil
	}

	return absolutePath, nil

}
