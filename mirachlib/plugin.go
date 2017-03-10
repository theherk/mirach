package mirachlib

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os/exec"

	"gitlab.eng.cleardata.com/dash/mirach/util"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/theherk/viper"
)

// ChunkSize defines the maximum size of chunks to be sent over MQTT.
const ChunkSize = 120000

// ExternalPlugin is a regularly run command that collects data.
type ExternalPlugin struct {
	Cmd      string `json:"cmd"`
	Label    string `json:"label"`
	Schedule string `json:"schedule"`
}

// InternalPlugin is a regularly run function that collects data.
type InternalPlugin struct {
	Label    string
	Schedule string
	StrFunc  func() string
	Type     string
}

type mqttMsg struct {
	Type string `json:"type"`
}

type chunksMsg struct {
	mqttMsg
	NumChunks int    `json:"num_chunks"`
	ChunksID  string `json:"chunks_id"`
	ChunksSum string `json:"chunks_sum"`
}

type dataMsg struct {
	mqttMsg
	Data json.RawMessage `json:"data"`
}

type urlMsg struct {
	mqttMsg
	URL string `json:"url"`
}

// Run will run an external plugin and publishes its results.
func (p *ExternalPlugin) Run(c mqtt.Client) func() {
	cmd := p.Cmd
	label := p.Label
	return func() {
		jww.INFO.Printf("Running external plugin: %s", label)
		cmd := exec.Command(cmd)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			jww.ERROR.Println(err)
		}
		if err := cmd.Start(); err != nil {
			jww.ERROR.Println(err)
		}
		var d []byte
		if err := json.NewDecoder(stdout).Decode(&d); err != nil {
			jww.ERROR.Println(err)
		}
		err = cmd.Wait()
		if err := SendData([]byte(d), c, label); err != nil {
			jww.ERROR.Println(err)
		}
	}
}

// Run will run an internal function and publish its results.
func (p *InternalPlugin) Run(c mqtt.Client) func() {
	f := p.StrFunc
	label := p.Label
	t := p.Type
	return func() {
		jww.INFO.Printf("Running internal plugin: %s", label)
		d := f()
		if err := SendData([]byte(d), c, t); err != nil {
			jww.ERROR.Println(err)
		}
	}
}

func postData(b []byte) (string, error) {
	// TODO: This will implement the push to S3.
	return "theurl", nil
}

// SendData sends data using one of a few methods to an MQTT broker.
func SendData(b []byte, c mqtt.Client, t string) error {
	custID := viper.GetString("customer.id")
	assetID := viper.GetString("asset.id")
	var err error
	msg := mqttMsg{Type: t}
	var msgB []byte
	switch {
	// TODO: case when implementing large storage upload
	// case len(b) >= 512000:
	// 	data, err := postData(b)
	// 	if err != nil {
	// 		return err
	// 	}
	case len(b) >= ChunkSize:
		n, id, sum, err := SendChunks(b, c)
		if err != nil {
			return err
		}
		m := chunksMsg{msg, n, id, sum}
		msgB, err = json.Marshal(m)
		if err != nil {
			return err
		}
	default:
		m := dataMsg{msg, json.RawMessage(string(b))}
		msgB, err = json.Marshal(m)
		if err != nil {
			return err
		}
	}
	path := fmt.Sprintf("mirach/data/%s/%s", custID, assetID)
	if err := PubWait(c, path, msgB); err != nil {
		return err
	}
	return nil
}

// SendChunks splits a byte slice and return the number of chunks, ID of
// the chunk group, hex digest of the full byte slice's md5 hash, or any
// errors generated along the way.
func SendChunks(b []byte, c mqtt.Client) (int, string, string, error) {
	id := fmt.Sprintf("%s", uuid.New())
	h := md5.Sum(b)
	sum := hex.EncodeToString(h[:])
	var n int
	splits, err := util.SplitAt(b, ChunkSize)
	if err != nil {
		return 0, "", "", err
	}
	for i, split := range splits {
		path := fmt.Sprintf("mirach/chunk/%s-%s", id, string(i))
		if err := PubWait(c, path, split); err != nil {
			return 0, "", "", err
		}
		n++
	}
	return n, id, sum, nil
}
