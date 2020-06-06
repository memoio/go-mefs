package fsrepo

import (
	"os"
	"path/filepath"

	config "github.com/memoio/go-mefs/config"
	id "github.com/memoio/go-mefs/crypto/identity"
	"github.com/memoio/go-mefs/utils/address"
)

// PutPrivateKeyToKeystore puts to keystore
func PutPrivateKeyToKeystore(privatekey string, peerID string, password string) error {
	dir := GetKeyStore()

	peerAddr, err := address.GetAddressFromID(peerID)
	if err != nil {
		return err
	}

	return id.StorePrivateKey(dir, privatekey, peerAddr.String(), password)
}

//GetPrivateKeyFromKeystore 是从keystore中获得ethereum格式的私钥, 没有"0x"前缀
func GetPrivateKeyFromKeystore(peerID string, password string) (privateKey string, err error) {
	//get privatekey's filepath, the dafault is "~/.mefs/keystore/peerID"
	dir := GetKeyStore()
	filePath, err := config.Path(dir, peerID)
	if err != nil {
		return "", err
	}

	_, err = os.Lstat(filePath)
	if os.IsNotExist(err) {
		// get peeraddress file
		peerAddr, err := address.GetAddressFromID(peerID)
		if err != nil {
			return "", err
		}

		filePath, err = config.Path(dir, peerAddr.String())
		if err != nil {
			return "", err
		}
		_, err = os.Lstat(filePath)
		if os.IsNotExist(err) {
			return "", err
		}
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

func GetKeyStore() string {
	fsrepoPath, _ := BestKnownPath()

	dir, _ := config.Path(fsrepoPath, Keystore)
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		err = os.Mkdir(dir, 0755)
		if err != nil {
			return ""
		}
	}

	return dir
}
