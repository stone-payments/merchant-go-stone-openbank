package openbank

import (
	"encoding/json"
	"fmt"
	"gopkg.in/square/go-jose.v2"
	"io/ioutil"
)

const (
	stonePublicKeysEndpoint = `/api/v1/discovery/keys`
)

func (c *Client) DecryptAndValidateWebhook(encryptedData string) (json.RawMessage, error) {
	decryptedData, err := c.DecryptJWE(encryptedData)
	if err != nil {
		return nil, fmt.Errorf(`failed at decrypting webhook: %w`, err)
	}

	jwe, err := jose.ParseSigned(string(decryptedData))
	if err != nil {
		return nil, fmt.Errorf(`err parsing webhook data: %w`, err)
	}

	signatureKey := c.getSignatureKey(jwe.Signatures)
	if _, err := jwe.Verify(signatureKey); err != nil {
		return nil, fmt.Errorf(`err verifying webhook signature: %w`, err)
	}

	return nil, nil
}

func (c *Client) getSignatureKey(signatures []jose.Signature) *jose.JSONWebKey {
	if len(signatures) != 1 {
		return nil // multi signature not supported
	}
	signature := signatures[0]
	return c.getStonePublicKey(signature.Header.KeyID)
}

func (c *Client) getStonePublicKey(id string) *jose.JSONWebKey {
	if key := c.StonePublicKeys.Get(id); key != nil {
		return key
	}

	if err := c.refreshPublicKeys(); err != nil {
		c.log.WithError(err).Error(`failure refreshing stone public keys`)
		return nil
	}

	return c.StonePublicKeys.Get(id)
}

func (c *Client) refreshPublicKeys() error {
	keysURL, err := c.ApiBaseURL.Parse(stonePublicKeysEndpoint)
	if err != nil {
		return fmt.Errorf(`failure parsing endpoint: %w`, err)
	}

	response, err := c.client.Get(keysURL.String())
	if err != nil {
		return err
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	type responseBody struct {
		Keys []jose.JSONWebKey `json:"keys"`
	}
	var r responseBody
	if err = json.Unmarshal(body, &r); err != nil {
		return err
	}

	for _, k := range r.Keys {
		c.StonePublicKeys[k.KeyID] = &k
	}

	return nil
}

func (c *Client) DecryptJWE(encryptedBody string) ([]byte, error) {
	jwe, err := jose.ParseEncrypted(encryptedBody)
	if err != nil {
		return nil, err
	}

	data, err := jwe.Decrypt(c.privateKey)
	if err != nil {
		return nil, err
	}

	return data, nil
}
