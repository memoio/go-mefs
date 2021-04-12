package commands

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/memoio/go-mefs/core/commands/cmdenv"

	cmds "github.com/ipfs/go-ipfs-cmds"
	id "github.com/memoio/go-mefs/crypto/identity"
	"github.com/memoio/go-mefs/utils"
)

var SkCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline: "Export the private key from the local keystore file",
		ShortDescription: `
Export the private key from the local keystore file, the file path must be provided.
The path can be absolute path or relative path.
For example: /home/zl/.mefs/keystore/0x2F34Aae01b7A66502d114EbcC50b16C78E645C32
`,
	},
	Arguments: []cmds.Argument{
		cmds.StringArg("path", true, false, "The absolute path of the keystore file."),
	},
	Options: []cmds.Option{
		cmds.StringOption("password", "pwd", "the password is used to decrypt the keystore file").WithDefault(utils.DefaultPassword),
	},

	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		n, err := cmdenv.GetNode(env)
		if err != nil {
			return err
		}

		password := req.Options["password"].(string)

		if n.Identity == "" {
			return errors.New("Identity is not loaded, please run daemon first")
		}
		peerID := n.Identity.Pretty()
		fmt.Println("this node's peerID is: ", peerID)

		filePath := req.Arguments[0]

		_, err = os.Lstat(filePath)
		if os.IsNotExist(err) {
			return err
		}

		//get config.PeerID from MefsNode.Identity
		sk, err := id.GetPrivateKey(peerID, password, filePath)
		if err != nil {
			return err
		}

		list := &StringList{
			ChildLists: []string{"private key : " + sk},
		}
		return cmds.EmitOnce(res, list)
	},
	Encoders: cmds.EncoderMap{
		cmds.Text: cmds.MakeTypedEncoder(func(req *cmds.Request, w io.Writer, fl *StringList) error {
			_, err := fmt.Fprintf(w, "%s", fl)
			return err
		}),
	},
	Type: IdOutput{},
}
