package control

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"testing"
)

func TestPasskeyRegistrationAndAssertionVerification(t *testing.T) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	rpID := "localhost"
	origin := "http://localhost:18080"
	challenge := []byte("challenge-123")
	credentialID := []byte("credential-123")
	publicKeyCOSE := testCOSEES256PublicKey(t, privateKey)

	registrationPayload := passkeyCredentialResponse{
		ID:    encodeBase64URL(credentialID),
		RawID: encodeBase64URL(credentialID),
		Type:  "public-key",
	}
	registrationClientData := testClientDataJSON(t, "webauthn.create", challenge, origin)
	registrationPayload.Response.ClientDataJSON = encodeBase64URL(registrationClientData)
	registrationPayload.Response.AttestationObject = encodeBase64URL(testAttestationObject(t, rpID, credentialID, publicKeyCOSE))
	registrationPayload.Response.Transports = []string{"internal"}

	credential, gotChallenge, err := parsePasskeyRegistration(registrationPayload, rpID, origin)
	if err != nil {
		t.Fatalf("parse registration: %v", err)
	}
	if string(gotChallenge) != string(challenge) {
		t.Fatalf("registration challenge = %q, want %q", gotChallenge, challenge)
	}
	if string(credential.CredentialID) != string(credentialID) {
		t.Fatalf("credential id = %q, want %q", credential.CredentialID, credentialID)
	}
	if credential.SignCount != 1 || credential.AttestationType != "none" {
		t.Fatalf("bad registered credential: %#v", credential)
	}

	assertionPayload := passkeyCredentialResponse{
		ID:    encodeBase64URL(credentialID),
		RawID: encodeBase64URL(credentialID),
		Type:  "public-key",
	}
	assertionClientData := testClientDataJSON(t, "webauthn.get", challenge, origin)
	assertionAuthData := testAssertionAuthData(rpID, 2)
	signatureBase := append([]byte(nil), assertionAuthData...)
	clientDataHash := sha256.Sum256(assertionClientData)
	signatureBase = append(signatureBase, clientDataHash[:]...)
	signatureDigest := sha256.Sum256(signatureBase)
	signature, err := ecdsa.SignASN1(rand.Reader, privateKey, signatureDigest[:])
	if err != nil {
		t.Fatal(err)
	}
	assertionPayload.Response.ClientDataJSON = encodeBase64URL(assertionClientData)
	assertionPayload.Response.AuthenticatorData = encodeBase64URL(assertionAuthData)
	assertionPayload.Response.Signature = encodeBase64URL(signature)

	assertionData, gotChallenge, err := verifyPasskeyAssertion(assertionPayload, credential.PublicKeyCOSE, credential.SignCount, rpID, origin)
	if err != nil {
		t.Fatalf("verify assertion: %v", err)
	}
	if string(gotChallenge) != string(challenge) {
		t.Fatalf("assertion challenge = %q, want %q", gotChallenge, challenge)
	}
	if assertionData.SignCount != 2 {
		t.Fatalf("assertion sign count = %d, want 2", assertionData.SignCount)
	}
}

func testClientDataJSON(t *testing.T, typ string, challenge []byte, origin string) []byte {
	t.Helper()
	data, err := json.Marshal(webauthnClientData{
		Type:      typ,
		Challenge: encodeBase64URL(challenge),
		Origin:    origin,
	})
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func testAttestationObject(t *testing.T, rpID string, credentialID, publicKeyCOSE []byte) []byte {
	t.Helper()
	return testCBORMapText(map[string][]byte{
		"fmt":      testCBORText("none"),
		"authData": testCBORBytes(testRegistrationAuthData(rpID, credentialID, publicKeyCOSE)),
		"attStmt":  testCBORMapRaw(nil),
	})
}

func testRegistrationAuthData(rpID string, credentialID, publicKeyCOSE []byte) []byte {
	rpIDHash := sha256.Sum256([]byte(rpID))
	out := append([]byte(nil), rpIDHash[:]...)
	out = append(out, 0x41)
	out = binary.BigEndian.AppendUint32(out, 1)
	out = append(out, make([]byte, 16)...)
	out = binary.BigEndian.AppendUint16(out, uint16(len(credentialID)))
	out = append(out, credentialID...)
	out = append(out, publicKeyCOSE...)
	return out
}

func testAssertionAuthData(rpID string, signCount uint32) []byte {
	rpIDHash := sha256.Sum256([]byte(rpID))
	out := append([]byte(nil), rpIDHash[:]...)
	out = append(out, 0x01)
	out = binary.BigEndian.AppendUint32(out, signCount)
	return out
}

func testCOSEES256PublicKey(t *testing.T, key *ecdsa.PrivateKey) []byte {
	t.Helper()
	x := key.PublicKey.X.Bytes()
	y := key.PublicKey.Y.Bytes()
	xPadded := append(make([]byte, 32-len(x)), x...)
	yPadded := append(make([]byte, 32-len(y)), y...)
	return testCBORMapRaw([][]byte{
		append(testCBORInt(1), testCBORInt(2)...),
		append(testCBORInt(3), testCBORInt(-7)...),
		append(testCBORInt(-1), testCBORInt(1)...),
		append(testCBORInt(-2), testCBORBytes(xPadded)...),
		append(testCBORInt(-3), testCBORBytes(yPadded)...),
	})
}

func testCBORMapText(values map[string][]byte) []byte {
	entries := make([][]byte, 0, len(values))
	for key, value := range values {
		entries = append(entries, append(testCBORText(key), value...))
	}
	return testCBORMapRaw(entries)
}

func testCBORMapRaw(entries [][]byte) []byte {
	out := testCBORMajor(5, uint64(len(entries)))
	for _, entry := range entries {
		out = append(out, entry...)
	}
	return out
}

func testCBORText(value string) []byte {
	return append(testCBORMajor(3, uint64(len(value))), []byte(value)...)
}

func testCBORBytes(value []byte) []byte {
	return append(testCBORMajor(2, uint64(len(value))), value...)
}

func testCBORInt(value int64) []byte {
	if value >= 0 {
		return testCBORMajor(0, uint64(value))
	}
	return testCBORMajor(1, uint64(-1-value))
}

func testCBORMajor(major byte, value uint64) []byte {
	prefix := major << 5
	switch {
	case value < 24:
		return []byte{prefix | byte(value)}
	case value <= 0xff:
		return []byte{prefix | 24, byte(value)}
	case value <= 0xffff:
		return []byte{prefix | 25, byte(value >> 8), byte(value)}
	default:
		return []byte{prefix | 26, byte(value >> 24), byte(value >> 16), byte(value >> 8), byte(value)}
	}
}
