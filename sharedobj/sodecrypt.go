
package main

import (
	"os"
	"fmt"
	"strings"
	"sync"
	"time"
	"path/filepath"
	"golang.org/x/crypto/openpgp/packet"

	session "github.com/aws/aws-sdk-go/aws/session"

	log "github.com/sirupsen/logrus"

    // local
	zip "github.com/tempuslabs/s3s2/zip"
	encrypt "github.com/tempuslabs/s3s2/encrypt"
	manifest "github.com/tempuslabs/s3s2/manifest"
	options "github.com/tempuslabs/s3s2/options"
	aws_helpers "github.com/tempuslabs/s3s2/aws_helpers"
	utils "github.com/tempuslabs/s3s2/utils"
	file "github.com/tempuslabs/s3s2/file"
)
import "C"

var opts options.Options

 //export Decrypt
func Decrypt(
	bucket string,
	f string,
	directory string,
	org string,
	region string,
	awsProfile string,
	privKey string,
	pubKey string,
	ssmPrivKey string,
	ssmPubKey string,
	isGCS bool,
	parallelism int) int {

	opts := options.Options{
		Bucket:      bucket,
		File:        f,
		Directory:   directory,
		Org:         org,
		Region:      region,
		PrivKey:     privKey,
		PubKey:      pubKey,
		IsGCS: 		 isGCS,
		SSMPrivKey:  ssmPrivKey,
		SSMPubKey:   ssmPubKey,
		AwsProfile:  awsProfile,
		Parallelism: parallelism,
	}
	checkDecryptOptions(opts)

	// top level clients
	sess := utils.GetAwsSession(opts)
	_pubKey := encrypt.GetPubKey(sess, opts)
	_privKey := encrypt.GetPrivKey(sess, opts)

	os.MkdirAll(opts.Directory, os.ModePerm)

	// if downloading via manifest
	if strings.HasSuffix(opts.File, "manifest.json") {

		log.Info("Detected manifest file...")

		target_manifest_path := filepath.Join(opts.Directory, filepath.Base(opts.File))
		fn, err := aws_helpers.DownloadFile(sess, opts.Bucket, opts.Org, opts.File, target_manifest_path, opts)
		utils.PanicIfError("Unable to download file at strings.HasSuffix - ", err)

		m := manifest.ReadManifest(fn)
		batch_folder := m.Folder
		file_structs := m.Files

		var wg sync.WaitGroup
		sem := make(chan int, opts.Parallelism)

		for _, fs := range file_structs {
			wg.Add(1)
			go func(wg *sync.WaitGroup, sess *session.Session, _pubkey *packet.PublicKey, _privKey *packet.PrivateKey, folder string, fs file.File, opts options.Options) {
				sem <- 1
				defer func() { <-sem }()
				defer wg.Done()
				// if block is for cases where AWS session expires, so we re-create session and attempt file again
				if decryptFile(sess, _pubKey, _privKey, m, fs, opts) != nil {
					sess = utils.GetAwsSession(opts)
					err := decryptFile(sess, _pubKey, _privKey, m, fs, opts)
					if err != nil {}
						log.Warn("Error during decrypt-file session expiration if block!")
						log.Errorf("Error: '%v'", err)
						panic(err)
				}
			}(&wg, sess, _pubKey, _privKey, batch_folder, fs, opts)
		}
		wg.Wait()
	}
	return 1
}

func decryptFile(sess *session.Session, _pubkey *packet.PublicKey, _privkey *packet.PrivateKey, m manifest.Manifest, fs file.File, opts options.Options) error {
	start := time.Now()
	log.Debugf("Starting decryption on file '%s'", fs.Name)

	// enforce posix path
	fs.Name = utils.ToPosixPath(fs.Name)

	aws_key := fs.GetEncryptedName(m.Folder)
	target_path := fs.GetEncryptedName(opts.Directory)

	fn_zip := fs.GetZipName(opts.Directory)
	fn_decrypt := fs.GetSourceName("decrypted")

	nested_dir := filepath.Dir(target_path)
	os.MkdirAll(nested_dir, os.ModePerm)

	_, err := aws_helpers.DownloadFile(sess, opts.Bucket, m.Organization, aws_key, target_path, opts)
	utils.PanicIfError("Main download failed - ", err)

    encrypt.DecryptFile(_pubkey, _privkey, target_path, fn_zip, opts)
	zip.UnZipFile(fn_zip, fn_decrypt, opts.Directory)

    utils.Timing(start, fmt.Sprintf("\tProcessed file '%s' in ", fs.Name) + "%f seconds")

	return err
}

func checkDecryptOptions(options options.Options) {
	if options.File == "" {
		log.Warn("Need to supply a file to decrypt. Should be the file path within the bucket but not including the bucket.")
		log.Panic("Insufficient information to perform decryption.")
	} else if options.Bucket == "" {
		log.Warn("Need to supply a bucket.")
		log.Panic("Insufficient information to perform decryption.")
	} else if options.Directory == "" {
		log.Warn("Need to supply a destination for the files to decrypt.  Should be a local path.")
		log.Panic("Insufficient information to perform decryption.")
	} else if options.Region == "" {
		log.Warn("Need to supply a region for the S3 bucket.")
		log.Panic("Insufficient information to perform decryption.")
	} else if options.PubKey == "" && options.SSMPubKey == "" {
	    log.Warn("Need to supply a public encryption key parameter.")
		log.Panic("Insufficient information to perform decryption.")
	} else if options.PrivKey == "" && options.SSMPrivKey == "" {
	    log.Warn("Need to supply a private encryption key parameter.")
		log.Panic("Insufficient information to perform decryption.")
	}
}

func main() {

}