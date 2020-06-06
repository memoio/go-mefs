package fsrepo

import (
	"os"

	config "github.com/memoio/go-mefs/config"
	homedir "github.com/mitchellh/go-homedir"
)

// BestKnownPath returns the best known fsrepo path. If the ENV override is
// present, this function returns that value. Otherwise, it returns the default
// repo path.
func BestKnownPath() (string, error) {
	mefsPath := config.DefaultPathRoot
	if os.Getenv(config.EnvDir) != "" { //获取环境变量
		mefsPath = os.Getenv(config.EnvDir)
	}
	mefsPath, err := homedir.Expand(mefsPath)
	if err != nil {
		return "", err
	}

	_, err = os.Stat(mefsPath)
	if os.IsNotExist(err) {
		err = os.Mkdir(mefsPath, 0755)
		if err != nil {
			return "", err
		}
	}
	return mefsPath, nil
}
