package commands

import (
	"fmt"
	"io"

	cmds "github.com/ipfs/go-ipfs-cmds"
	id "github.com/memoio/go-mefs/crypto/identity"
	fsrepo "github.com/memoio/go-mefs/repo/fsrepo"
	"github.com/memoio/go-mefs/utils"
)

const (
	SecreteKey = "secreteKey"
	PassWord   = "password"
)

type UserPrivMessage struct {
	Address string
	Sk      string
}

var CreateCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "create user.",
		ShortDescription: `
   create a MEFS user by a specified Sk or generate a new sk.
   you can use a MEFS service in a trusted node(agent).
`,
	},
	Arguments: []cmds.Argument{},
	Options: []cmds.Option{
		cmds.StringOption(SecreteKey, "sk", "The practice user's privatekey that you want to create").WithDefault(""),
		cmds.StringOption(PassWord, "pwd", "The practice user's password that you want to exec").WithDefault(utils.DefaultPassword),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		pwd, found := req.Options[PassWord].(string)
		if pwd == "" || !found {
			pwd = utils.DefaultPassword
		}
		sk, found := req.Options[SecreteKey].(string)
		if sk == "" || !found {
			tsk, err := id.Create()
			if err != nil {
				return err
			}
			sk = id.ECDSAByteToString(id.ToECDSAByte(tsk))
		}
		pub, err := id.GetPubByte(sk)
		if err != nil {
			return err
		}
		pid, err := id.GetIDFromPubKey(pub)
		if err != nil {
			return err
		}

		err = fsrepo.PutPrivateKeyToKeystore(sk, pid, pwd)
		if err != nil {
			return err
		}

		addr, err := id.GetAdressFromSk(sk)
		if err != nil {
			return err
		}

		return cmds.EmitOnce(res, &UserPrivMessage{Address: addr.String(), Sk: sk})
	},
	Type: UserPrivMessage{},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeTypedEncoder(func(req *cmds.Request, w io.Writer, up *UserPrivMessage) error {
			_, err := fmt.Fprintf(w, "Private Key: %s\nAddress: %s\n", up.Sk, up.Address)
			return err
		}),
	},
}
