package openbank

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/apex/log"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/stone-co/go-stone-openbank/types"
	"gopkg.in/square/go-jose.v2"
)

func (c *Client) WebhookFromRequest(r *http.Request) (*types.WebhookPayload, error) {
	var payload struct {
		EncryptedBody string `json:"encrypted_body"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return nil, err
	}

	return c.Webhook(payload.EncryptedBody)

}

func (c *Client) Webhook(key string) (*types.WebhookPayload, error) {

	signBytes, err := ioutil.ReadFile(c.PrivateKeyPath)
	if err != nil {
		return nil, err
	}
	privKey, err := jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	if err != nil {
		return nil, err
	}

	jwe, err := jose.ParseEncrypted(key)
	if err != nil {
		return nil, err
	}

	data, err := jwe.Decrypt(privKey)
	if err != nil {
		return nil, err
	}

	token, err := jose.ParseSigned(string(data))
	if err != nil {
		return nil, err
	}
	log.Infof("%+v", token)

	//	jwt.ParseSigned(data)
	//
	//JWS
	//set, err := jwk.Fetch("https://sandbox-api.openbank.stone.com.br/api/v1/discovery/keys")
	//	if err != nil {
	//		return nil, err
	//	}
	//
	//	keyID, ok := token.Header["kid"].(string)
	//	if !ok {
	//		return nil, errors.New("expecting JWT header to have string kid")
	//	}
	//
	//	if key := set.LookupKeyID(keyID); len(key) == 1 {
	//		return key[0].Materialize()
	//	}
	//
	//	//	verificationKey, err := LoadJSONWebKey([]byte(`{e: "AQAB",
	//	//kid: "03eaae04-9ef2-41ac-885d-d01deeafbcb2",
	//	//kty: "RSA",
	//	//n: "oy7_hdhkYezmKFVek3dwhpD6wKNro20Ben6TXeAS4LQj-8qb9e5rWrkChWqN4Xc-dxq_cp5IP_Uz-BaW10EMIC2VfeAXQuR98rktpgUyYnthwIuLeZoz3OKS8dKppY_UrKlYfXODs2ak9EM8o8Bjw6t0-NO4xlw0-pvqGKWzCXe0Qq8t1hi609-fU8N9M_z8jcuMEGHEcemxrAxaEbiB5kNeHn6N4R3FT7I7TKGr1HVk30mZfQsjGYjV-B-sjDzTEmrTeUKDzogzvG78u4hFDewREiodcA7Vq2u9yWO_LK17n_6j0UkReZ4b9r1WPAsGGajasK7S1urcVyvL97TYq70SW54qNuliSlbI3L44yHredVOd1rdR3ifZin-u5we7MKq9YkMgD48P4MPLdMuX26ef4991erK5_ksHaqz-mVACQvHC0_uhucEqdwTodarZvq0Ro1KjVLfmbVvwErL5AqNzy-mIdPuikWlilSg_Oaka2P5HPp-9kspOawZhAuwU7U_xFsvxcpiKG47pLj8eJn-PJbdAGdFDucsVhWGMyFKlIA3oqZMikYLUFvXeBqntPN1bC-Fj_9ycPHabkYPbvUQgvau-zAnfGGo-Yh7dQTHWAcIqdnyZT1mZwHNBmUpzw5uKwbGUs3tuIXvhDnPNRxiofJvYa6JwDjk8xA8mkx0",
	//	//use: "sig"}`), true)
	//	//	if err != nil {
	//	//		return nil, err
	//	//	}
	//	//
	//	//	payload, err := obj.Verify(verificationKey)
	//	//	if err != nil {
	//	//		return nil, err
	//	//	}
	//	log.Infof("%s", payload)

	return nil, nil
}

func LoadJSONWebKey(json []byte, pub bool) (*jose.JSONWebKey, error) {
	var jwk jose.JSONWebKey
	err := jwk.UnmarshalJSON(json)
	if err != nil {
		return nil, err
	}
	if !jwk.Valid() {
		return nil, errors.New("invalid JWK key")
	}
	if jwk.IsPublic() != pub {
		return nil, errors.New("priv/pub JWK key mismatch")
	}
	return &jwk, nil
}
