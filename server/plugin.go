package main

import (
	"bytes"
	"errors"
	"fmt"
	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
	"golang.org/x/crypto/openpgp/packet"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"

	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin"
)

// Plugin implements the interface expected by the Mattermost server to communicate between the server and plugin processes.
type Plugin struct {
	plugin.MattermostPlugin

	// configurationLock synchronizes access to the configuration.
	configurationLock sync.RWMutex

	// configuration is the active plugin configuration. Consult getConfiguration and
	// setConfiguration for usage.
	configuration *configuration
}

// ServeHTTP demonstrates a plugin that handles HTTP requests by greeting the world.
func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello, world!")
}

// FileWillBeUploaded hook
func (p *Plugin) FileWillBeUploaded(c *plugin.Context, info *model.FileInfo, file io.Reader, output io.Writer) (*model.FileInfo, string) {
	data, err := ioutil.ReadAll(file)
	if err != nil {
		p.API.LogError(err.Error())
		return nil, err.Error()
	}

	// passphrase will be moved to env variable in the future.
	passphrase := []byte("myTemporaryPassphrase1")

	packetConfig := &packet.Config{
		DefaultCipher: packet.CipherAES256,
	}

	encryptedData, _ := Encrypt(data, passphrase, packetConfig)

	if _, err := output.Write([]byte(encryptedData)); err != nil {
		p.API.LogError(err.Error())
	}

	return nil, ""
}

// Encrypt function source: https://play.golang.org/p/vk58yYArMh and https://asecuritysite.com/encryption/go_pgp
func Encrypt(plaintext []byte, password []byte, packetConfig *packet.Config) (ciphertext []byte, err error) {
	encbuf := bytes.NewBuffer(nil)

	w, _ := armor.Encode(encbuf, "PGP MESSAGE", nil)
	pt, _ := openpgp.SymmetricallyEncrypt(w, password, nil, packetConfig)
	_, err = pt.Write(plaintext)
	if err != nil {
		return
	}
	pt.Close()
	w.Close()
	ciphertext = encbuf.Bytes()
	return
}

// FileWillBeRead hook (we need to implement this hook. I couldn't yet.)
func (p *Plugin) FileWillBeRead(c *plugin.Context, file io.Reader) {
	data, err := ioutil.ReadAll(file)
	if err != nil {
		p.API.LogError(err.Error())
		return
	}

	// passphrase will be moved to env variable in the future.
	passphrase := []byte("myTemporaryPassphrase1")

	packetConfig := &packet.Config{
		DefaultCipher: packet.CipherAES256,
	}

	decryptedData, _ := Decrypt(data, passphrase, packetConfig)
	// fmt.Printf(string(decrypted))

	file = strings.NewReader(string(decryptedData))
	return
}

// Decrypt function source: https://play.golang.org/p/vk58yYArMh and https://asecuritysite.com/encryption/go_pgp
func Decrypt(ciphertext []byte, password []byte, packetConfig *packet.Config) (plaintext []byte, err error) {
	decbuf := bytes.NewBuffer(ciphertext)
	armorBlock, _ := armor.Decode(decbuf)
	failed := false
	prompt := func(keys []openpgp.Key, symmetric bool) ([]byte, error) {
		if failed {
			return nil, errors.New("decryption failed")
		}
		failed = true
		return password, nil
	}
	md, err := openpgp.ReadMessage(armorBlock.Body, nil, prompt, packetConfig)
	if err != nil {
		return
	}
	plaintext, err = ioutil.ReadAll(md.UnverifiedBody)
	if err != nil {
		return
	}
	return
}
