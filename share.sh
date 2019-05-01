#!/bin/bash

aws-vault exec jemurai-mkonda-admin -- go run main.go share --debug true --bucket s3s2-demo --region us-east-1 --directory test/s3s2/s3s2-up/ --org Jemurai --prefix jemurai --receiver-public-key test/s3s2/s3s2-keys/test.pubkey
