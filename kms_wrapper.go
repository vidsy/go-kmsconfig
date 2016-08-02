package aws

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
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
func (k KMSWrapper) Decrypt(ciphertextBlob string) (string, error) {
	output, err := k.Client.Decrypt(k.decryptParmas(ciphertextBlob))

	if err != nil {
		return "", err
	}

	return string(output.Plaintext[:]), nil
}

func (k KMSWrapper) decryptParmas(ciphertextBlob string) *kms.DecryptInput {
	return &kms.DecryptInput{
		CiphertextBlob: []byte(ciphertextBlob),
	}
}
