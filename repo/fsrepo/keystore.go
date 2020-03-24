package fsrepo

import (
	"path/filepath"

	config "github.com/memoio/go-mefs/config"
	id "github.com/memoio/go-mefs/crypto/identity"
)

// PutPrivateKeyToKeystore puts to keystore
func PutPrivateKeyToKeystore(privatekey string, peerID string, password string) error {
	rootpath, _ := BestKnownPath()
	keypath, _ := config.Path(rootpath, Keystore)

	return id.StorePrivateKey(keypath, privatekey, peerID, password)
}

//GetPrivateKeyFromKeystore 是从keystore中获得ethereum格式的私钥, 没有"0x"前缀
func GetPrivateKeyFromKeystore(peerID string, password string) (privateKey string, err error) {
	//get privatekey's filepath, the dafault is "~/.ipfs/keystore/peerID"
	fsrepoPath, err := BestKnownPath()
	dir, err := config.Path(fsrepoPath, Keystore)
	filePath, err := config.Path(dir, peerID)
	if err != nil {
		return "", err
	}

	//get config.PeerID from MefsNode.Identity
	key, err := id.GetPrivateKey(peerID, password, filePath)
	if err != nil {
		return "", err
	}

	return key, nil
}

func joinPath(dir string, filename string) (path string) {
	if filepath.IsAbs(filename) {
		return filename
	}
	return filepath.Join(dir, filename)
}
