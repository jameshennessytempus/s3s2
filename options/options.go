package options

// Options is the information we need about a particular sharing activity.
type Options struct {
	// For both encrypt/decrypt
	Region string `json:"region"`
	Bucket string `json:"bucket"`
	AwsProfile string `json:"awsprofile"`

	// Encrypt only
	PubKey    string `json:"pubkey"`
	SSMPubKey string `json:"ssmpubkey"`
	Directory string `json:"directory"`
	AwsKey    string `json:"awskey"`
	Org       string `json:"org"`
	Prefix    string `json:"prefix"`
	Hash      bool   `json:"hash"`

	// Decrypt only
	File        string `json:"file"`
	Destination string `json:"destination"`
	PrivKey     string `json:"privkey"`
	SSMPrivKey string `json:"ssmprivkey"`
}
