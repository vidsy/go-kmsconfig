package kmsconfig

import (
	"encoding/base64"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"log"
)

type (
	// KMSWrapper comment pending
	KMSWrapper struct {
		Client *kms.KMS
	}
)

// NewKMSWrapper comment pending
func NewKMSWrapper() KMSWrapper {
	return KMSWrapper{
		kms.New(session.New()),
	}
}

// Decrypt comment pending
func (k KMSWrapper) Decrypt(encodedCipherTextBlob string) (string, error) {
	decodedValue, err := base64.StdEncoding.DecodeString(encodedCipherTextBlob)

	if err != nil {
		log.Printf("Could not Base64 decode: '%s'", encodedCipherTextBlob)
		return "", err
	}

	output, err := k.Client.Decrypt(k.decryptParmas(decodedValue))

	if err != nil {
		return "", err
	}

	return string(output.Plaintext[:]), nil
}

func (k KMSWrapper) decryptParmas(cipherTextBlob []byte) *kms.DecryptInput {
	return &kms.DecryptInput{
		CiphertextBlob: cipherTextBlob,
	}
}
