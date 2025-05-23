package fsrepo

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	lockfile "github.com/ipfs/go-fs-lock"
	util "github.com/ipfs/go-ipfs-util"
	logging "github.com/ipfs/go-log"
	config "github.com/memoio/go-mefs/config"
	serialize "github.com/memoio/go-mefs/config/serialize"
	repo "github.com/memoio/go-mefs/repo"
	"github.com/memoio/go-mefs/repo/common"
	ds "github.com/memoio/go-mefs/source/go-datastore"
	measure "github.com/memoio/go-mefs/source/go-ds-measure"
	"github.com/memoio/go-mefs/utils"
	homedir "github.com/mitchellh/go-homedir"
	ma "github.com/multiformats/go-multiaddr"
)

// LockFile is the filename of the repo lock, relative to config dir
// TODO rename repo lock and hide name
const (
	LockFile               = "repo.lock"
	nBitsForKeypairDefault = 2048
	version                = 3
)

var log = logging.Logger("fsrepo")

// version number that we are currently expecting to see
var RepoVersion = 1

var (
	ErrNoVersion = errors.New("version file is wrong.")
	ErrKSExist   = errors.New("mefs keystore already exists")
)

type NoRepoError struct {
	Path string
}

var _ error = NoRepoError{}

func (err NoRepoError) Error() string {
	return fmt.Sprintf("no MEFS repo found in %s.\nplease run: 'mefs init'", err.Path)
}

const apiFile = "api"
const swarmKeyFile = "swarm.key"

const specFn = "datastore_spec"
const Keystore = "keystore"

const VersionFile = "version"

var (

	// packageLock must be held to while performing any operation that modifies an
	// FSRepo's state field. This includes Init, Open, Close, and Remove.
	packageLock sync.Mutex

	// onlyOne keeps track of open FSRepo instances.
	//
	// TODO: once command Context / Repo integration is cleaned up,
	// this can be removed. Right now, this makes ConfigCmd.Run
	// function try to open the repo twice:
	//
	//     $ mefs daemon &
	//     $ mefs config foo
	//
	// The reason for the above is that in standalone mode without the
	// daemon, `mefs config` tries to save work by not building the
	// full MefsNode, but accessing the Repo directly.
	onlyOne repo.OnlyOne
)

// FSRepo represents an MEFS FileSystem Repo. It is safe for use by multiple
// callers.
type FSRepo struct {
	// has Close been called already
	closed bool
	// path is the file-system path
	path string
	// lockfile is the file system lock to prevent others from opening
	// the same fsrepo path concurrently
	lockfile io.Closer
	config   *config.Config
	ds       repo.Datastore
}

var _ repo.Repo = (*FSRepo)(nil)

// Open the FSRepo at path. Returns an error if the repo is not
// initialized.
func Open(repoPath string) (repo.Repo, error) {
	fn := func() (repo.Repo, error) {
		return open(repoPath)
	}
	return onlyOne.Open(repoPath, fn)
}

func open(repoPath string) (repo.Repo, error) {
	packageLock.Lock()
	defer packageLock.Unlock()

	r, err := newFSRepo(repoPath)
	if err != nil {
		return nil, err
	}

	// Check if its initialized
	if err := checkInitialized(r.path); err != nil {
		return nil, err
	}

	r.lockfile, err = lockfile.Lock(r.path, LockFile)
	if err != nil {
		return nil, err
	}
	keepLocked := false
	defer func() {
		// unlock on error, leave it locked on success
		if !keepLocked {
			r.lockfile.Close()
		}
	}()

	// check repo path, then check all constituent parts.
	if err := utils.Writable(r.path); err != nil {
		return nil, err
	}

	if err := r.openConfig(); err != nil {
		return nil, err
	}

	if err := r.openDatastore(); err != nil {
		return nil, err
	}

	keepLocked = true
	return r, nil
}

func newFSRepo(rpath string) (*FSRepo, error) {
	expPath, err := homedir.Expand(filepath.Clean(rpath))
	if err != nil {
		return nil, err
	}

	return &FSRepo{path: expPath}, nil
}

func checkInitialized(path string) error {
	if !isInitializedUnsynced(path) {
		alt := strings.Replace(path, ".mefs", ".go-mefs", 1)
		if isInitializedUnsynced(alt) {
			return ErrNoVersion
		}
		return NoRepoError{Path: path}
	}
	return nil
}

// ConfigAt returns an error if the FSRepo at the given path is not
// initialized. This function allows callers to read the config file even when
// another process is running and holding the lock.
func ConfigAt(repoPath string) (*config.Config, error) {

	// packageLock must be held to ensure that the Read is atomic.
	packageLock.Lock()
	defer packageLock.Unlock()

	configFilename, err := config.Filename(repoPath)
	if err != nil {
		return nil, err
	}
	return serialize.Load(configFilename)
}

// configIsInitialized returns true if the repo is initialized at
// provided |path|.
func configIsInitialized(path string) bool {
	configFilename, err := config.Filename(path)
	if err != nil {
		return false
	}
	if !util.FileExists(configFilename) {
		return false
	}
	return true
}

func initConfig(path string, conf *config.Config) error {
	if configIsInitialized(path) {
		return nil
	}
	configFilename, err := config.Filename(path)
	if err != nil {
		return err
	}
	// initialization is the one time when it's okay to write to the config
	// without reading the config from disk and merging any user-provided keys
	// that may exist.
	if err := serialize.WriteConfigFile(configFilename, conf); err != nil {
		return err
	}

	return nil
}

func initVersion(rp string, verison int) error {
	fn := path.Join(string(rp), VersionFile)
	return ioutil.WriteFile(fn, []byte(fmt.Sprintf("%d\n", version)), 0644)
}

func initSpec(path string, conf map[string]interface{}) error {
	fn, err := config.Path(path, specFn)
	if err != nil {
		return err
	}

	if util.FileExists(fn) {
		return nil
	}

	dsc, err := AnyDatastoreConfig(conf)
	if err != nil {
		return err
	}
	bytes := dsc.DiskSpec().Bytes()

	return ioutil.WriteFile(fn, bytes, 0600)
}

func initKeyStore(path, privateKey, peerID, password string) error {
	dir, err := config.Path(path, Keystore)
	if err != nil {
		return err
	}

	_, err = os.Stat(dir)
	// dir not exist
	if os.IsNotExist(err) {
		err = os.Mkdir(dir, 0755)
		if err != nil {
			return err
		}
	}

	return PutPrivateKeyToKeystore(privateKey, peerID, password)
}

// Init initializes a new FSRepo at the given path with the provided config.
// TODO add support for custom datastores.
func Init(repoPath string, conf *config.Config, prikey string, password string) error {

	// packageLock must be held to ensure that the repo is not initialized more
	// than once.
	packageLock.Lock()
	defer packageLock.Unlock()

	if isInitializedUnsynced(repoPath) {
		return nil
	}

	if err := initConfig(repoPath, conf); err != nil {
		return err
	}

	if err := initSpec(repoPath, conf.Datastore.Spec); err != nil {
		return err
	}

	if err := initVersion(repoPath, RepoVersion); err != nil {
		return err
	}

	if err := initKeyStore(repoPath, prikey, conf.PeerID, password); err != nil {
		return err
	}

	return nil
}

// LockedByOtherProcess returns true if the FSRepo is locked by another
// process. If true, then the repo cannot be opened by this process.
func LockedByOtherProcess(repoPath string) (bool, error) {
	repoPath = filepath.Clean(repoPath)
	locked, err := lockfile.Locked(repoPath, LockFile)
	if locked {
		log.Debugf("(%t)<->Lock is held at %s", locked, repoPath)
	}
	return locked, err
}

// APIAddr returns the registered API addr, according to the api file
// in the fsrepo. This is a concurrent operation, meaning that any
// process may read this file. modifying this file, therefore, should
// use "mv" to replace the whole file and avoid interleaved read/writes.
func APIAddr(repoPath string) (ma.Multiaddr, error) {
	repoPath = filepath.Clean(repoPath)
	apiFilePath := filepath.Join(repoPath, apiFile)

	// if there is no file, assume there is no api addr.
	f, err := os.Open(apiFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, repo.ErrApiNotRunning
		}
		return nil, err
	}
	defer f.Close()

	// read up to 2048 bytes. io.ReadAll is a vulnerability, as
	// someone could hose the process by putting a massive file there.
	//
	// NOTE(@stebalien): @jbenet probably wasn't thinking straight when he
	// wrote that comment but I'm leaving the limit here in case there was
	// some hidden wisdom. However, I'm fixing it such that:
	// 1. We don't read too little.
	// 2. We don't truncate and succeed.
	buf, err := ioutil.ReadAll(io.LimitReader(f, 2048))
	if err != nil {
		return nil, err
	}
	if len(buf) == 2048 {
		return nil, fmt.Errorf("API file too large, must be <2048 bytes long: %s", apiFilePath)
	}

	s := string(buf)
	s = strings.TrimSpace(s)
	return ma.NewMultiaddr(s)
}

func (r *FSRepo) Path() string {
	return r.path
}

// SetAPIAddr writes the API Addr to the /api file.
func (r *FSRepo) SetAPIAddr(addr ma.Multiaddr) error {
	f, err := os.Create(filepath.Join(r.path, apiFile))
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(addr.String())
	return err
}

// openConfig returns an error if the config file is not present.
func (r *FSRepo) openConfig() error {
	configFilename, err := config.Filename(r.path)
	if err != nil {
		return err
	}
	conf, err := serialize.Load(configFilename)
	if err != nil {
		return err
	}
	r.config = conf
	return nil
}

// openDatastore returns an error if the config file is not present.
func (r *FSRepo) openDatastore() error {
	if r.config.Datastore.Type != "" || r.config.Datastore.Path != "" {
		return fmt.Errorf("old style datatstore config detected")
	} else if r.config.Datastore.Spec == nil {
		return fmt.Errorf("required Datastore.Spec entry missing from config file")
	}
	if r.config.Datastore.NoSync {
		log.Warning("NoSync is now deprecated in favor of datastore specific settings. If you want to disable fsync on flatfs set 'sync' to false. See https://github.com/memoio/go-mefs/blob/master/docs/datastores.md#flatfs.")
	}

	dsc, err := AnyDatastoreConfig(r.config.Datastore.Spec)
	if err != nil {
		return err
	}
	spec := dsc.DiskSpec()

	oldSpec, err := r.readSpec()
	if err != nil {
		return err
	}
	if oldSpec != spec.String() {
		return fmt.Errorf("datastore configuration of '%s' does not match what is on disk '%s'",
			oldSpec, spec.String())
	}

	d, err := dsc.Create(r.path)
	if err != nil {
		return err
	}
	r.ds = d

	// Wrap it with metrics gathering
	prefix := "mefs.fsrepo.datastore"
	r.ds = measure.New(prefix, r.ds)

	return nil
}

func (r *FSRepo) readSpec() (string, error) {
	fn, err := config.Path(r.path, specFn)
	if err != nil {
		return "", err
	}
	b, err := ioutil.ReadFile(fn)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}

// Close closes the FSRepo, releasing held resources.
func (r *FSRepo) Close() error {
	packageLock.Lock()
	defer packageLock.Unlock()

	if r.closed {
		return errors.New("repo is closed")
	}

	err := os.Remove(filepath.Join(r.path, apiFile))
	if err != nil && !os.IsNotExist(err) {
		log.Warning("error removing api file: ", err)
	}

	if err := r.ds.Close(); err != nil {
		return err
	}

	// This code existed in the previous versions, but
	// EventlogComponent.Close was never called. Preserving here
	// pending further discussion.
	//
	// TODO It isn't part of the current contract, but callers may like for us
	// to disable logging once the component is closed.
	// logging.Configure(logging.Output(os.Stderr))

	r.closed = true
	return r.lockfile.Close()
}

// Result when not Open is undefined. The method may panic if it pleases.
func (r *FSRepo) Config() (*config.Config, error) {

	// It is not necessary to hold the package lock since the repo is in an
	// opened state. The package lock is _not_ meant to ensure that the repo is
	// thread-safe. The package lock is only meant to guard against removal and
	// coordinate the lockfile. However, we provide thread-safety to keep
	// things simple.
	packageLock.Lock()
	defer packageLock.Unlock()

	if r.closed {
		return nil, errors.New("cannot access config, repo not open")
	}
	return r.config, nil
}

func (r *FSRepo) BackupConfig(prefix string) (string, error) {
	temp, err := ioutil.TempFile(r.path, "config-"+prefix)
	if err != nil {
		return "", err
	}
	defer temp.Close()

	configFilename, err := config.Filename(r.path)
	if err != nil {
		return "", err
	}

	orig, err := os.OpenFile(configFilename, os.O_RDONLY, 0600)
	if err != nil {
		return "", err
	}
	defer orig.Close()

	_, err = io.Copy(temp, orig)
	if err != nil {
		return "", err
	}

	return orig.Name(), nil
}

// setConfigUnsynced is for private use.
func (r *FSRepo) setConfigUnsynced(updated *config.Config) error {
	configFilename, err := config.Filename(r.path)
	if err != nil {
		return err
	}
	// to avoid clobbering user-provided keys, must read the config from disk
	// as a map, write the updated struct values to the map and write the map
	// to disk.
	var mapconf map[string]interface{}
	if err := serialize.ReadConfigFile(configFilename, &mapconf); err != nil {
		return err
	}
	m, err := config.ToMap(updated)
	if err != nil {
		return err
	}
	for k, v := range m {
		mapconf[k] = v
	}
	if err := serialize.WriteConfigFile(configFilename, mapconf); err != nil {
		return err
	}
	*r.config = *updated // copy so caller cannot modify this private config
	return nil
}

// SetConfig updates the FSRepo's config.
func (r *FSRepo) SetConfig(updated *config.Config) error {

	// packageLock is held to provide thread-safety.
	packageLock.Lock()
	defer packageLock.Unlock()

	return r.setConfigUnsynced(updated)
}

// GetConfigKey retrieves only the value of a particular key.
func (r *FSRepo) GetConfigKey(key string) (interface{}, error) {
	packageLock.Lock()
	defer packageLock.Unlock()

	if r.closed {
		return nil, errors.New("repo is closed")
	}

	filename, err := config.Filename(r.path)
	if err != nil {
		return nil, err
	}
	var cfg map[string]interface{}
	if err := serialize.ReadConfigFile(filename, &cfg); err != nil {
		return nil, err
	}
	return common.MapGetKV(cfg, key)
}

// SetConfigKey writes the value of a particular key.
func (r *FSRepo) SetConfigKey(key string, value interface{}) error {
	packageLock.Lock()
	defer packageLock.Unlock()

	if r.closed {
		return errors.New("repo is closed")
	}

	filename, err := config.Filename(r.path)
	if err != nil {
		return err
	}
	var mapconf map[string]interface{}
	if err := serialize.ReadConfigFile(filename, &mapconf); err != nil {
		return err
	}

	// Get the type of the value associated with the key
	oldValue, err := common.MapGetKV(mapconf, key)
	ok := true
	if err != nil {
		// key-value does not exist yet
		switch v := value.(type) {
		case string:
			value, err = strconv.ParseBool(v)
			if err != nil {
				value, err = strconv.Atoi(v)
				if err != nil {
					value, err = strconv.ParseFloat(v, 32)
					if err != nil {
						value = v
					}
				}
			}
		default:
		}
	} else {
		switch oldValue.(type) {
		case bool:
			value, ok = value.(bool)
		case int:
			value, ok = value.(int)
		case float32:
			value, ok = value.(float32)
		case string:
			value, ok = value.(string)
		default:
		}
		if !ok {
			return fmt.Errorf("wrong config type, expected %T", oldValue)
		}
	}

	if err := common.MapSetKV(mapconf, key, value); err != nil {
		return err
	}

	// This step doubles as to validate the map against the struct
	// before serialization
	conf, err := config.FromMap(mapconf)
	if err != nil {
		return err
	}
	if err := serialize.WriteConfigFile(filename, mapconf); err != nil {
		return err
	}
	return r.setConfigUnsynced(conf) // TODO roll this into this method
}

// Datastore returns a repo-owned datastore. If FSRepo is Closed, return value
// is undefined.
func (r *FSRepo) Datastore() repo.Datastore {
	packageLock.Lock()
	d := r.ds
	packageLock.Unlock()
	return d
}

// GetStorageUsage computes the storage space taken by the repo in bytes
func (r *FSRepo) GetStorageUsage() (uint64, error) {
	return ds.DiskUsage(r.Datastore())
}

func (r *FSRepo) SwarmKey() ([]byte, error) {
	repoPath := filepath.Clean(r.path)
	spath := filepath.Join(repoPath, swarmKeyFile)

	f, err := os.Open(spath)
	if err != nil {
		if os.IsNotExist(err) {
			err = nil
		}
		return nil, err
	}
	defer f.Close()

	return ioutil.ReadAll(f)
}

var _ io.Closer = &FSRepo{}
var _ repo.Repo = &FSRepo{}

// IsInitialized returns true if the repo is initialized at provided |path|.
func IsInitialized(path string) bool {
	// packageLock is held to ensure that another caller doesn't attempt to
	// Init or Remove the repo while this call is in progress.
	packageLock.Lock()
	defer packageLock.Unlock()

	return isInitializedUnsynced(path)
}

// private methods below this point. NB: packageLock must held by caller.

// isInitializedUnsynced reports whether the repo is initialized. Caller must
// hold the packageLock.
func isInitializedUnsynced(repoPath string) bool {
	return configIsInitialized(repoPath)
}
