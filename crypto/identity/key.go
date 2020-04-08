package key

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	cr "crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	btcec "github.com/btcsuite/btcd/btcec"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
	ci "github.com/libp2p/go-libp2p-core/crypto"
	peer "github.com/libp2p/go-libp2p-core/peer"
	"golang.org/x/crypto/scrypt"
)

//Key the struct privatekey transform to
type Key struct {
	ID uuid.UUID
	// to simplify lookups we also store the address
	PeerID string
	// we only store privkey as pubkey/address can be derived from it
	// privkey in this struct is always in plaintext
	// Eth format: 66bytes
	PrivateKey []byte
}

type cipherparamsJSON struct {
	IV string `json:"iv"`
}

type cryptoJSON struct {
	Cipher       string                 `json:"cipher"`
	CipherText   string                 `json:"ciphertext"`
	CipherParams cipherparamsJSON       `json:"cipherparams"`
	KDF          string                 `json:"kdf"`
	KDFParams    map[string]interface{} `json:"kdfparams"`
	MAC          string                 `json:"mac"`
}

type encryptedKeyJSONV3 struct {
	Address string     `json:"address"`
	Crypto  cryptoJSON `json:"crypto"`
	ID      string     `json:"id"`
	Version int        `json:"version"`
}

const (
	keyHeaderKDF = "scrypt"
	// StandardScryptN is the N parameter of Scrypt encryption algorithm, using 256MB
	// memory and taking approximately 1s CPU time on a modern processor.
	StandardScryptN = 1 << 18
	// StandardScryptP is the P parameter of Scrypt encryption algorithm, using 256MB
	// memory and taking approximately 1s CPU time on a modern processor.
	StandardScryptP = 1
	scryptR         = 8
	scryptDKLen     = 32
	version         = 3
)

var (
	//ErrDecrypt before decrypt privatekey, we compare mac, if not equal, use ErrDecrypt
	ErrDecrypt = errors.New("could not decrypt key with given passphrase")
)

func newKey(privatekey string, peerID string) (*Key, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	keyBytes, err := ECDSAStringToByte(privatekey)
	if err != nil {
		return nil, err
	}
	key := &Key{
		ID:         id,
		PeerID:     peerID,
		PrivateKey: keyBytes,
	}
	return key, nil
}

//StorePrivateKey encrypt the privatekey by password and then store it in keystore
func StorePrivateKey(dir, privatekey, peerID, password string) error {
	path := joinPath(dir, peerID)
	_, err := os.Stat(path)
	if os.IsExist(err) {
		return nil
	}
	key, err := newKey(privatekey, peerID)
	if err != nil {
		return err
	}

	keyjson, err := encryptKey(key, password, StandardScryptN, StandardScryptP)
	if err != nil {
		return err
	}

	return writeKeyFile(path, keyjson) //写入文件
}

func GetPrivateKey(peerID, password, path string) (privateKey string, err error) {
	// Load the key from the keystore and decrypt its contents
	keyjson, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	key, err := decryptKey(keyjson, password)
	if err != nil {
		return "", err
	}
	// Make sure we're really operating on the requested key (no swap attacks)
	if peerID != "" && key.PeerID != peerID {
		return "", fmt.Errorf("key content mismatch: have peer %x, want %x", key.PeerID, peerID)
	}

	return ECDSAByteToString(key.PrivateKey), nil
}

func joinPath(dir string, filename string) (path string) {
	if filepath.IsAbs(filename) {
		return filename
	}
	return filepath.Join(dir, filename)
}

// EncryptKey encrypts a key using the specified scrypt parameters into a json
// blob that can be decrypted later on.
func encryptKey(key *Key, password string, scryptN, scryptP int) ([]byte, error) {
	passwordArray := []byte(password)
	salt := getEntropyCSPRNG(32)                                                               //生成一个随即的32B的salt
	derivedKey, err := scrypt.Key(passwordArray, salt, scryptN, scryptR, scryptP, scryptDKLen) //使用scrypt算法对输入的password加密，生成一个32位的derivedKey
	if err != nil {
		return nil, err
	}
	encryptKey := derivedKey[:16] //why

	iv := getEntropyCSPRNG(aes.BlockSize)                        // 16,aes-128-ctr加密算法需要的初始化向量
	cipherText, err := aesCTRXOR(encryptKey, key.PrivateKey, iv) //对privatekey进行aes加密，生成一个32byte的cipherText
	if err != nil {
		return nil, err
	}
	mac := crypto.Keccak256(derivedKey[16:32], cipherText) //将derivedKey的后16byte与cipherText进行Keccak256哈希，生成32byte的mac，mac用于验证解密时password的正确性

	scryptParamsJSON := make(map[string]interface{}, 5)
	scryptParamsJSON["n"] = scryptN
	scryptParamsJSON["r"] = scryptR
	scryptParamsJSON["p"] = scryptP
	scryptParamsJSON["dklen"] = scryptDKLen
	scryptParamsJSON["salt"] = hex.EncodeToString(salt)

	cipherParamsJSON := cipherparamsJSON{
		IV: hex.EncodeToString(iv),
	}

	cryptoStruct := cryptoJSON{
		Cipher:       "aes-128-ctr",
		CipherText:   hex.EncodeToString(cipherText),
		CipherParams: cipherParamsJSON,
		KDF:          keyHeaderKDF,
		KDFParams:    scryptParamsJSON,
		MAC:          hex.EncodeToString(mac),
	}
	encryptedKeyJSONV3 := encryptedKeyJSONV3{
		key.PeerID,
		cryptoStruct,
		key.ID.String(),
		version,
	}
	return json.Marshal(encryptedKeyJSONV3)
}

func getEntropyCSPRNG(n int) []byte {
	mainBuff := make([]byte, n)
	_, err := io.ReadFull(cr.Reader, mainBuff)
	if err != nil {
		panic("reading from crypto/rand failed: " + err.Error())
	}
	return mainBuff
}

func aesCTRXOR(key, inText, iv []byte) ([]byte, error) {
	// AES-128 is selected due to size of encryptKey.
	aesBlock, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	stream := cipher.NewCTR(aesBlock, iv)
	outText := make([]byte, len(inText))
	stream.XORKeyStream(outText, inText)
	return outText, err
}

func writeKeyFile(file string, content []byte) error {
	// Atomic write: create a temporary hidden file first
	// then move it into place. TempFile assigns mode 0600.
	//mode 0600 represents that the owner has read and write permission and no execution permission.
	f, err := ioutil.TempFile(filepath.Dir(file), "."+filepath.Base(file)+".tmp")
	if err != nil {
		return err
	}
	if _, err := f.Write(content); err != nil {
		f.Close()
		os.Remove(f.Name())
		return err
	}
	f.Close()
	return os.Rename(f.Name(), file)
}

// decryptKey decrypts a key from a json blob, returning the private key itself.
func decryptKey(keyjson []byte, password string) (*Key, error) {
	// Depending on the version try to parse one way or another
	var (
		keyBytes []byte
		keyID    [16]byte
		err      error
	)
	k := new(encryptedKeyJSONV3)
	if err := json.Unmarshal(keyjson, k); err != nil {
		return nil, err
	}
	keyBytes, keyID, err = decryptKeyV3(k, password)
	// Handle any decryption errors and return the key
	if err != nil {
		return nil, err
	}
	key := crypto.ToECDSAUnsafe(keyBytes)

	if key == nil {
		return nil, ErrDecrypt
	}

	secpkey := (*ci.Secp256k1PrivateKey)((*btcec.PrivateKey)(key))
	pubkey := secpkey.GetPublic()
	id, err := peer.IDFromPublicKey(pubkey)
	return &Key{
		ID:         uuid.UUID(keyID),
		PeerID:     id.Pretty(),
		PrivateKey: keyBytes,
	}, nil
}

func decryptKeyV3(keyProtected *encryptedKeyJSONV3, password string) (keyBytes []byte, keyID [16]byte, err error) {
	if keyProtected.Version != version {
		return nil, keyID, fmt.Errorf("Version not supported: %v", keyProtected.Version)
	}

	if keyProtected.Crypto.Cipher != "aes-128-ctr" {
		return nil, keyID, fmt.Errorf("Cipher not supported: %v", keyProtected.Crypto.Cipher)
	}

	keyID, err = uuid.Parse(keyProtected.ID)
	mac, err := hex.DecodeString(keyProtected.Crypto.MAC)
	if err != nil {
		return nil, keyID, err
	}

	iv, err := hex.DecodeString(keyProtected.Crypto.CipherParams.IV)
	if err != nil {
		return nil, keyID, err
	}

	cipherText, err := hex.DecodeString(keyProtected.Crypto.CipherText)
	if err != nil {
		return nil, keyID, err
	}

	derivedKey, err := getKDFKey(keyProtected.Crypto, password)
	if err != nil {
		return nil, keyID, err
	}

	calculatedMAC := crypto.Keccak256(derivedKey[16:32], cipherText)
	if !bytes.Equal(calculatedMAC, mac) {
		return nil, keyID, ErrDecrypt
	}

	plainText, err := aesCTRXOR(derivedKey[:16], cipherText, iv)
	if err != nil {
		return nil, keyID, err
	}
	return plainText, keyID, err
}

func getKDFKey(cryptoJSON cryptoJSON, password string) ([]byte, error) {
	passwordArray := []byte(password)
	salt, err := hex.DecodeString(cryptoJSON.KDFParams["salt"].(string))
	if err != nil {
		return nil, err
	}
	dkLen := ensureInt(cryptoJSON.KDFParams["dklen"])

	if cryptoJSON.KDF == keyHeaderKDF {
		n := ensureInt(cryptoJSON.KDFParams["n"])
		r := ensureInt(cryptoJSON.KDFParams["r"])
		p := ensureInt(cryptoJSON.KDFParams["p"])
		return scrypt.Key(passwordArray, salt, n, r, p, dkLen)

	}
	return nil, fmt.Errorf("Unsupported KDF: %s", cryptoJSON.KDF)
}

func ensureInt(x interface{}) int {
	res, ok := x.(int)
	if !ok {
		res = int(x.(float64))
	}
	return res
}
