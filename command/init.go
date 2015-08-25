package command

import (
	"fmt"
	"strings"

	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/helper/pgpkeys"
)

// InitCommand is a Command that initializes a new Vault server.
type InitCommand struct {
	Meta
}

func (c *InitCommand) Run(args []string) int {
	var threshold, shares int
	var pgpKeys pgpkeys.PubKeyFilesFlag
	flags := c.Meta.FlagSet("init", FlagSetDefault)
	flags.Usage = func() { c.Ui.Error(c.Help()) }
	flags.IntVar(&shares, "key-shares", 5, "")
	flags.IntVar(&threshold, "key-threshold", 3, "")
	flags.Var(&pgpKeys, "pgp-keys", "")
	if err := flags.Parse(args); err != nil {
		return 1
	}

	client, err := c.Client()
	if err != nil {
		c.Ui.Error(fmt.Sprintf(
			"Error initializing client: %s", err))
		return 1
	}

	resp, err := client.Sys().Init(&api.InitRequest{
		SecretShares:    shares,
		SecretThreshold: threshold,
		PGPKeys:         pgpKeys,
	})
	if err != nil {
		c.Ui.Error(fmt.Sprintf(
			"Error initializing Vault: %s", err))
		return 1
	}

	for i, key := range resp.Keys {
		c.Ui.Output(fmt.Sprintf("Key %d: %s", i+1, key))
	}

	c.Ui.Output(fmt.Sprintf("Initial Root Token: %s", resp.RootToken))

	c.Ui.Output(fmt.Sprintf(
		"\n"+
			"Vault initialized with %d keys and a key threshold of %d. Please\n"+
			"securely distribute the above keys. When the Vault is re-sealed,\n"+
			"restarted, or stopped, you must provide at least %d of these keys\n"+
			"to unseal it again.\n\n"+
			"Vault does not store the master key. Without at least %d keys,\n"+
			"your Vault will remain permanently sealed.",
		shares,
		threshold,
		threshold,
		threshold,
	))

	return 0
}

func (c *InitCommand) Synopsis() string {
	return "Initialize a new Vault server"
}

func (c *InitCommand) Help() string {
	helpText := `
Usage: vault init [options]

  Initialize a new Vault server.

  This command connects to a Vault server and initializes it for the
  first time. This sets up the initial set of master keys and sets up the
  backend data store structure.

  This command can't be called on an already-initialized Vault.

General Options:

  ` + generalOptionsUsage() + `

Init Options:

  -key-shares=5           The number of key shares to split the master key
                          into.

  -key-threshold=3        The number of key shares required to reconstruct
                          the master key.

  -pgp-keys               If provided, must be a comma-separated list of
                          files on disk containing binary-format public PGP
                          keys. The number of files must match 'key-shares'.
                          The output unseal keys will be hex-encoded and
                          encrypted, in order, with the given public keys.
                          If you want to use them with the 'vault unseal'
                          command, you will need to hex decode, decrypt, and
                          hex encode the result; this will be the plaintext
                          unseal key.
`
	return strings.TrimSpace(helpText)
}
