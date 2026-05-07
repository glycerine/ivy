package control

import (
	"bytes"
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const passkeyChallengeTTL = 10 * time.Minute

type passkeyCredentialDescriptor struct {
	Type       string   `json:"type"`
	ID         string   `json:"id"`
	Transports []string `json:"transports,omitempty"`
}

type passkeyRelyingParty struct {
	Name string `json:"name"`
	ID   string `json:"id,omitempty"`
}

type passkeyUserEntity struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
}

type passkeyCredentialParam struct {
	Type string `json:"type"`
	Alg  int    `json:"alg"`
}

type passkeyAuthenticatorSelection struct {
	ResidentKey        string `json:"residentKey"`
	RequireResidentKey bool   `json:"requireResidentKey"`
	UserVerification   string `json:"userVerification"`
}

type passkeyRegistrationOptions struct {
	Challenge              string                        `json:"challenge"`
	RP                     passkeyRelyingParty           `json:"rp"`
	User                   passkeyUserEntity             `json:"user"`
	PubKeyCredParams       []passkeyCredentialParam      `json:"pubKeyCredParams"`
	Timeout                int                           `json:"timeout"`
	Attestation            string                        `json:"attestation"`
	AuthenticatorSelection passkeyAuthenticatorSelection `json:"authenticatorSelection"`
	ExcludeCredentials     []passkeyCredentialDescriptor `json:"excludeCredentials,omitempty"`
}

type passkeyRegistrationOptionsResponse struct {
	PublicKey passkeyRegistrationOptions `json:"publicKey"`
}

type passkeyLoginOptions struct {
	Challenge        string                        `json:"challenge"`
	RPID             string                        `json:"rpId,omitempty"`
	Timeout          int                           `json:"timeout"`
	UserVerification string                        `json:"userVerification"`
	AllowCredentials []passkeyCredentialDescriptor `json:"allowCredentials,omitempty"`
}

type passkeyLoginOptionsResponse struct {
	PublicKey passkeyLoginOptions `json:"publicKey"`
}

type passkeyCredentialResponse struct {
	ID       string `json:"id"`
	RawID    string `json:"rawId"`
	Type     string `json:"type"`
	Response struct {
		ClientDataJSON    string   `json:"clientDataJSON"`
		AttestationObject string   `json:"attestationObject"`
		AuthenticatorData string   `json:"authenticatorData"`
		Signature         string   `json:"signature"`
		UserHandle        string   `json:"userHandle"`
		Transports        []string `json:"transports"`
	} `json:"response"`
}

type webauthnClientData struct {
	Type        string `json:"type"`
	Challenge   string `json:"challenge"`
	Origin      string `json:"origin"`
	CrossOrigin bool   `json:"crossOrigin"`
}

type parsedAuthenticatorData struct {
	RPIDHash       []byte
	Flags          byte
	SignCount      uint32
	AAGUID         string
	CredentialID   []byte
	PublicKeyCOSE  []byte
	BackupEligible bool
	BackedUp       bool
}

func (s *Server) handlePasskeyRegisterOptions(w http.ResponseWriter, r *http.Request) {
	view, _, ok := s.authenticatedSessionView(w, r)
	if !ok {
		return
	}
	origin, rpID, publicRPID, err := s.passkeyOriginAndRPID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid public base url"})
		return
	}
	challenge, err := randomBytes(32)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create passkey challenge"})
		return
	}
	now := time.Now().UTC()
	if err := s.store.CreatePasskeyChallenge(r.Context(), view.User.ID, "registration", challenge, rpID, origin, now, passkeyChallengeTTL); err != nil {
		s.logf("passkey_register_options challenge_failed user_id=%s error=%q", view.User.ID, err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to store passkey challenge"})
		return
	}
	existingIDs, err := s.store.PasskeyCredentialIDsForUser(r.Context(), view.User.ID)
	if err != nil {
		s.logf("passkey_register_options existing_credentials_failed user_id=%s error=%q", view.User.ID, err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load existing passkeys"})
		return
	}
	exclude := make([]passkeyCredentialDescriptor, 0, len(existingIDs))
	for _, id := range existingIDs {
		exclude = append(exclude, passkeyCredentialDescriptor{
			Type: "public-key",
			ID:   encodeBase64URL(id),
		})
	}
	displayName := view.User.DisplayName
	if strings.TrimSpace(displayName) == "" {
		displayName = view.User.Email
	}
	writeJSON(w, http.StatusOK, passkeyRegistrationOptionsResponse{
		PublicKey: passkeyRegistrationOptions{
			Challenge: encodeBase64URL(challenge),
			RP: passkeyRelyingParty{
				Name: "Ivy",
				ID:   publicRPID,
			},
			User: passkeyUserEntity{
				ID:          encodeBase64URL([]byte(view.User.ID)),
				Name:        view.User.Email,
				DisplayName: displayName,
			},
			PubKeyCredParams: []passkeyCredentialParam{
				{Type: "public-key", Alg: -7},
				{Type: "public-key", Alg: -257},
			},
			Timeout:     60000,
			Attestation: "none",
			AuthenticatorSelection: passkeyAuthenticatorSelection{
				ResidentKey:        "required",
				RequireResidentKey: true,
				UserVerification:   "preferred",
			},
			ExcludeCredentials: exclude,
		},
	})
}

func (s *Server) handlePasskeyRegisterFinish(w http.ResponseWriter, r *http.Request) {
	view, sessionToken, ok := s.authenticatedSessionView(w, r)
	if !ok {
		return
	}
	var payload passkeyCredentialResponse
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	origin, rpID, _, err := s.passkeyOriginAndRPID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid public base url"})
		return
	}
	credential, challenge, err := parsePasskeyRegistration(payload, rpID, origin)
	if err != nil {
		s.logf("passkey_register_finish invalid user_id=%s error=%q", view.User.ID, err)
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid passkey registration"})
		return
	}
	if err := s.store.ConsumePasskeyChallenge(r.Context(), view.User.ID, "registration", challenge, rpID, origin, time.Now().UTC()); err != nil {
		status := http.StatusBadRequest
		if !errors.Is(err, ErrPasskeyChallengeNotFound) {
			status = http.StatusInternalServerError
		}
		s.logf("passkey_register_finish challenge_rejected user_id=%s error=%q", view.User.ID, err)
		writeJSON(w, status, map[string]string{"error": "passkey challenge is invalid or expired"})
		return
	}
	credential.UserID = view.User.ID
	credential.DisplayName = view.User.Email
	if err := s.store.CreatePasskeyCredential(r.Context(), credential); err != nil {
		s.logf("passkey_register_finish store_failed user_id=%s error=%q", view.User.ID, err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to store passkey"})
		return
	}
	refreshed, err := s.store.SessionViewByToken(r.Context(), sessionToken, time.Now().UTC(), AppSessionTTL)
	if err != nil {
		s.logf("passkey_register_finish session_reload_failed user_id=%s error=%q", view.User.ID, err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load updated session"})
		return
	}
	writeJSON(w, http.StatusOK, refreshed)
}

func (s *Server) handlePasskeyLoginOptions(w http.ResponseWriter, r *http.Request) {
	origin, rpID, publicRPID, err := s.passkeyOriginAndRPID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid public base url"})
		return
	}
	challenge, err := randomBytes(32)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create passkey challenge"})
		return
	}
	now := time.Now().UTC()
	if err := s.store.CreatePasskeyChallenge(r.Context(), "", "login", challenge, rpID, origin, now, passkeyChallengeTTL); err != nil {
		s.logf("passkey_login_options challenge_failed error=%q", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to store passkey challenge"})
		return
	}
	writeJSON(w, http.StatusOK, passkeyLoginOptionsResponse{
		PublicKey: passkeyLoginOptions{
			Challenge:        encodeBase64URL(challenge),
			RPID:             publicRPID,
			Timeout:          60000,
			UserVerification: "preferred",
		},
	})
}

func (s *Server) handlePasskeyLoginFinish(w http.ResponseWriter, r *http.Request) {
	var payload passkeyCredentialResponse
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	origin, rpID, _, err := s.passkeyOriginAndRPID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid public base url"})
		return
	}
	rawID, err := decodeBase64URL(payload.RawID)
	if err != nil || len(rawID) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid passkey credential id"})
		return
	}
	credential, err := s.store.PasskeyCredentialByID(r.Context(), rawID)
	if errors.Is(err, ErrPasskeyCredentialNotFound) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "passkey is not registered"})
		return
	}
	if err != nil {
		s.logf("passkey_login_finish lookup_failed error=%q", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load passkey"})
		return
	}
	authData, challenge, err := verifyPasskeyAssertion(payload, credential.PublicKeyCOSE, credential.SignCount, rpID, origin)
	if err != nil {
		s.logf("passkey_login_finish invalid credential_id=%s error=%q", encodeBase64URL(rawID), err)
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid passkey assertion"})
		return
	}
	if err := s.store.ConsumePasskeyChallenge(r.Context(), "", "login", challenge, rpID, origin, time.Now().UTC()); err != nil {
		status := http.StatusBadRequest
		if !errors.Is(err, ErrPasskeyChallengeNotFound) {
			status = http.StatusInternalServerError
		}
		s.logf("passkey_login_finish challenge_rejected credential_id=%s error=%q", encodeBase64URL(rawID), err)
		writeJSON(w, status, map[string]string{"error": "passkey challenge is invalid or expired"})
		return
	}
	if err := s.store.UpdatePasskeyCredentialSignCount(r.Context(), rawID, authData.SignCount, time.Now().UTC()); err != nil {
		s.logf("passkey_login_finish sign_count_update_failed credential_id=%s error=%q", encodeBase64URL(rawID), err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update passkey"})
		return
	}
	sessionToken, err := s.createAppSessionForUser(r.Context(), credential.UserID, time.Now().UTC())
	if err != nil {
		s.logf("passkey_login_finish session_failed user_id=%s error=%q", credential.UserID, err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create app session"})
		return
	}
	setAppCookie(w, SessionCookieName, sessionToken, AppSessionTTL, s.cfg.CookieSecure)
	view, err := s.store.SessionViewByToken(r.Context(), sessionToken, time.Now().UTC(), AppSessionTTL)
	if err != nil {
		s.logf("passkey_login_finish session_reload_failed user_id=%s error=%q", credential.UserID, err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load app session"})
		return
	}
	writeJSON(w, http.StatusOK, view)
}

func (s *Server) authenticatedSessionView(w http.ResponseWriter, r *http.Request) (SessionView, string, bool) {
	cookie, err := r.Cookie(SessionCookieName)
	if errors.Is(err, http.ErrNoCookie) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "authentication required"})
		return SessionView{}, "", false
	}
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid session cookie"})
		return SessionView{}, "", false
	}
	view, err := s.store.SessionViewByToken(r.Context(), cookie.Value, time.Now().UTC(), AppSessionTTL)
	if errors.Is(err, ErrSessionNotFound) {
		clearCookie(w, SessionCookieName, s.cfg.CookieSecure)
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "authentication required"})
		return SessionView{}, "", false
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load session"})
		return SessionView{}, "", false
	}
	if !view.Authenticated || view.User == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "authentication required"})
		return SessionView{}, "", false
	}
	if view.CookieRefreshNeeded {
		setAppCookie(w, SessionCookieName, cookie.Value, AppSessionTTL, s.cfg.CookieSecure)
	}
	return view, cookie.Value, true
}

func (s *Server) passkeyOriginAndRPID(r *http.Request) (string, string, string, error) {
	base, err := url.Parse(s.publicBaseURL(r))
	if err != nil {
		return "", "", "", err
	}
	if base.Scheme == "" || base.Host == "" {
		return "", "", "", errors.New("public base url requires scheme and host")
	}
	host := strings.ToLower(base.Hostname())
	if host == "" {
		return "", "", "", errors.New("public base url requires hostname")
	}
	origin := base.Scheme + "://" + base.Host
	publicRPID := host
	if net.ParseIP(host) != nil {
		publicRPID = ""
	}
	return origin, host, publicRPID, nil
}

func (s *Server) createAppSessionForUser(ctx context.Context, userID string, now time.Time) (string, error) {
	sessionToken, err := RandomToken(32)
	if err != nil {
		return "", err
	}
	csrfToken, err := RandomToken(32)
	if err != nil {
		return "", err
	}
	if err := s.store.CreateAppSession(ctx, userID, sessionToken, csrfToken, now, AppSessionTTL, AppSessionTTL); err != nil {
		return "", err
	}
	return sessionToken, nil
}

func parsePasskeyRegistration(payload passkeyCredentialResponse, rpID, origin string) (StoredPasskeyCredential, []byte, error) {
	if payload.Type != "public-key" {
		return StoredPasskeyCredential{}, nil, errors.New("credential type must be public-key")
	}
	rawID, err := decodeBase64URL(payload.RawID)
	if err != nil {
		return StoredPasskeyCredential{}, nil, err
	}
	clientDataJSON, err := decodeBase64URL(payload.Response.ClientDataJSON)
	if err != nil {
		return StoredPasskeyCredential{}, nil, err
	}
	challenge, err := verifyWebAuthnClientData(clientDataJSON, "webauthn.create", origin)
	if err != nil {
		return StoredPasskeyCredential{}, nil, err
	}
	attestationObject, err := decodeBase64URL(payload.Response.AttestationObject)
	if err != nil {
		return StoredPasskeyCredential{}, nil, err
	}
	credential, err := parseAttestationObject(attestationObject, rpID)
	if err != nil {
		return StoredPasskeyCredential{}, nil, err
	}
	if len(rawID) > 0 && !bytes.Equal(rawID, credential.CredentialID) {
		return StoredPasskeyCredential{}, nil, errors.New("raw credential id does not match attested credential id")
	}
	if len(payload.Response.Transports) > 0 {
		credential.Transports = normalizeTransports(payload.Response.Transports)
	}
	return credential, challenge, nil
}

func parseAttestationObject(attestationObject []byte, rpID string) (StoredPasskeyCredential, error) {
	value, err := decodeCBOR(attestationObject)
	if err != nil {
		return StoredPasskeyCredential{}, err
	}
	m, ok := value.(cborMap)
	if !ok {
		return StoredPasskeyCredential{}, errors.New("attestation object is not a cbor map")
	}
	format, ok := cborTextValue(m, "fmt")
	if !ok {
		return StoredPasskeyCredential{}, errors.New("attestation format is missing")
	}
	authDataBytes, ok := cborBytesValue(m, "authData")
	if !ok {
		return StoredPasskeyCredential{}, errors.New("attestation authData is missing")
	}
	authData, err := parseAuthenticatorData(authDataBytes, rpID, true)
	if err != nil {
		return StoredPasskeyCredential{}, err
	}
	if err := validatePasskeyPublicKey(authData.PublicKeyCOSE); err != nil {
		return StoredPasskeyCredential{}, err
	}
	return StoredPasskeyCredential{
		CredentialID:    authData.CredentialID,
		PublicKeyCOSE:   authData.PublicKeyCOSE,
		SignCount:       authData.SignCount,
		BackupEligible:  authData.BackupEligible,
		BackedUp:        authData.BackedUp,
		AttestationType: format,
		AAGUID:          authData.AAGUID,
	}, nil
}

func verifyPasskeyAssertion(payload passkeyCredentialResponse, publicKeyCOSE []byte, storedSignCount uint32, rpID, origin string) (parsedAuthenticatorData, []byte, error) {
	if payload.Type != "public-key" {
		return parsedAuthenticatorData{}, nil, errors.New("credential type must be public-key")
	}
	clientDataJSON, err := decodeBase64URL(payload.Response.ClientDataJSON)
	if err != nil {
		return parsedAuthenticatorData{}, nil, err
	}
	challenge, err := verifyWebAuthnClientData(clientDataJSON, "webauthn.get", origin)
	if err != nil {
		return parsedAuthenticatorData{}, nil, err
	}
	authenticatorData, err := decodeBase64URL(payload.Response.AuthenticatorData)
	if err != nil {
		return parsedAuthenticatorData{}, nil, err
	}
	authData, err := parseAuthenticatorData(authenticatorData, rpID, false)
	if err != nil {
		return parsedAuthenticatorData{}, nil, err
	}
	if authData.SignCount != 0 && storedSignCount != 0 && authData.SignCount <= storedSignCount {
		return parsedAuthenticatorData{}, nil, errors.New("passkey sign count did not increase")
	}
	signature, err := decodeBase64URL(payload.Response.Signature)
	if err != nil {
		return parsedAuthenticatorData{}, nil, err
	}
	if err := verifyPasskeySignature(publicKeyCOSE, authenticatorData, clientDataJSON, signature); err != nil {
		return parsedAuthenticatorData{}, nil, err
	}
	return authData, challenge, nil
}

func verifyWebAuthnClientData(clientDataJSON []byte, expectedType, expectedOrigin string) ([]byte, error) {
	var clientData webauthnClientData
	if err := json.Unmarshal(clientDataJSON, &clientData); err != nil {
		return nil, err
	}
	if clientData.Type != expectedType {
		return nil, fmt.Errorf("client data type = %q", clientData.Type)
	}
	if clientData.Origin != expectedOrigin {
		return nil, fmt.Errorf("client data origin = %q", clientData.Origin)
	}
	if clientData.CrossOrigin {
		return nil, errors.New("cross-origin passkey ceremony rejected")
	}
	challenge, err := decodeBase64URL(clientData.Challenge)
	if err != nil {
		return nil, err
	}
	if len(challenge) == 0 {
		return nil, errors.New("client data challenge is empty")
	}
	return challenge, nil
}

func parseAuthenticatorData(data []byte, rpID string, requireAttestedCredential bool) (parsedAuthenticatorData, error) {
	if len(data) < 37 {
		return parsedAuthenticatorData{}, errors.New("authenticator data is too short")
	}
	rpIDHash := sha256.Sum256([]byte(rpID))
	if !bytes.Equal(data[:32], rpIDHash[:]) {
		return parsedAuthenticatorData{}, errors.New("rp id hash mismatch")
	}
	flags := data[32]
	if flags&0x01 == 0 {
		return parsedAuthenticatorData{}, errors.New("user presence flag is not set")
	}
	authData := parsedAuthenticatorData{
		RPIDHash:       append([]byte(nil), data[:32]...),
		Flags:          flags,
		SignCount:      binary.BigEndian.Uint32(data[33:37]),
		BackupEligible: flags&0x08 != 0,
		BackedUp:       flags&0x10 != 0,
	}
	if !requireAttestedCredential {
		return authData, nil
	}
	if flags&0x40 == 0 {
		return parsedAuthenticatorData{}, errors.New("attested credential data flag is not set")
	}
	offset := 37
	if len(data) < offset+18 {
		return parsedAuthenticatorData{}, errors.New("attested credential data is too short")
	}
	authData.AAGUID = uuidStringFromBytes(data[offset : offset+16])
	offset += 16
	credentialIDLen := int(binary.BigEndian.Uint16(data[offset : offset+2]))
	offset += 2
	if len(data) < offset+credentialIDLen {
		return parsedAuthenticatorData{}, errors.New("credential id overruns authenticator data")
	}
	authData.CredentialID = append([]byte(nil), data[offset:offset+credentialIDLen]...)
	offset += credentialIDLen
	keyLen, err := cborItemEnd(data[offset:])
	if err != nil {
		return parsedAuthenticatorData{}, err
	}
	authData.PublicKeyCOSE = append([]byte(nil), data[offset:offset+keyLen]...)
	return authData, nil
}

func verifyPasskeySignature(publicKeyCOSE, authenticatorData, clientDataJSON, signature []byte) error {
	cose, err := decodeCBOR(publicKeyCOSE)
	if err != nil {
		return err
	}
	m, ok := cose.(cborMap)
	if !ok {
		return errors.New("cose public key is not a map")
	}
	alg, ok := cborIntValue(m, 3)
	if !ok {
		return errors.New("cose public key missing alg")
	}
	signedBytes := make([]byte, 0, len(authenticatorData)+sha256.Size)
	signedBytes = append(signedBytes, authenticatorData...)
	clientDataHash := sha256.Sum256(clientDataJSON)
	signedBytes = append(signedBytes, clientDataHash[:]...)
	digest := sha256.Sum256(signedBytes)
	switch alg {
	case -7:
		publicKey, err := parseCOSEES256PublicKey(m)
		if err != nil {
			return err
		}
		if !ecdsa.VerifyASN1(publicKey, digest[:], signature) {
			return errors.New("ecdsa passkey signature verification failed")
		}
		return nil
	case -257:
		publicKey, err := parseCOSERS256PublicKey(m)
		if err != nil {
			return err
		}
		return rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, digest[:], signature)
	default:
		return fmt.Errorf("unsupported cose alg %d", alg)
	}
}

func validatePasskeyPublicKey(publicKeyCOSE []byte) error {
	cose, err := decodeCBOR(publicKeyCOSE)
	if err != nil {
		return err
	}
	m, ok := cose.(cborMap)
	if !ok {
		return errors.New("cose public key is not a map")
	}
	alg, ok := cborIntValue(m, 3)
	if !ok {
		return errors.New("cose public key missing alg")
	}
	switch alg {
	case -7:
		_, err := parseCOSEES256PublicKey(m)
		return err
	case -257:
		_, err := parseCOSERS256PublicKey(m)
		return err
	default:
		return fmt.Errorf("unsupported cose alg %d", alg)
	}
}

func parseCOSEES256PublicKey(m cborMap) (*ecdsa.PublicKey, error) {
	kty, ok := cborIntValue(m, 1)
	if !ok || kty != 2 {
		return nil, errors.New("cose ec2 public key has invalid kty")
	}
	crv, ok := cborIntValue(m, -1)
	if !ok || crv != 1 {
		return nil, errors.New("cose ec2 public key is not p-256")
	}
	xBytes, ok := cborBytesIntValue(m, -2)
	if !ok || len(xBytes) != 32 {
		return nil, errors.New("cose ec2 public key has invalid x coordinate")
	}
	yBytes, ok := cborBytesIntValue(m, -3)
	if !ok || len(yBytes) != 32 {
		return nil, errors.New("cose ec2 public key has invalid y coordinate")
	}
	curve := elliptic.P256()
	x := new(big.Int).SetBytes(xBytes)
	y := new(big.Int).SetBytes(yBytes)
	if !curve.IsOnCurve(x, y) {
		return nil, errors.New("cose ec2 public key point is not on curve")
	}
	return &ecdsa.PublicKey{Curve: curve, X: x, Y: y}, nil
}

func parseCOSERS256PublicKey(m cborMap) (*rsa.PublicKey, error) {
	kty, ok := cborIntValue(m, 1)
	if !ok || kty != 3 {
		return nil, errors.New("cose rsa public key has invalid kty")
	}
	nBytes, ok := cborBytesIntValue(m, -1)
	if !ok || len(nBytes) == 0 {
		return nil, errors.New("cose rsa public key has invalid modulus")
	}
	eBytes, ok := cborBytesIntValue(m, -2)
	if !ok || len(eBytes) == 0 || len(eBytes) > 8 {
		return nil, errors.New("cose rsa public key has invalid exponent")
	}
	exponent := new(big.Int).SetBytes(eBytes)
	if !exponent.IsInt64() || exponent.Int64() < 3 {
		return nil, errors.New("cose rsa public key exponent is invalid")
	}
	return &rsa.PublicKey{
		N: new(big.Int).SetBytes(nBytes),
		E: int(exponent.Int64()),
	}, nil
}

func randomBytes(n int) ([]byte, error) {
	if n <= 0 {
		return nil, errors.New("byte count must be positive")
	}
	out := make([]byte, n)
	if _, err := rand.Read(out); err != nil {
		return nil, err
	}
	return out, nil
}

func encodeBase64URL(raw []byte) string {
	return base64.RawURLEncoding.EncodeToString(raw)
}

func decodeBase64URL(encoded string) ([]byte, error) {
	encoded = strings.TrimSpace(encoded)
	if encoded == "" {
		return nil, nil
	}
	if decoded, err := base64.RawURLEncoding.DecodeString(encoded); err == nil {
		return decoded, nil
	}
	return base64.URLEncoding.DecodeString(encoded)
}

func uuidStringFromBytes(raw []byte) string {
	if len(raw) != 16 {
		return ""
	}
	allZero := true
	for _, b := range raw {
		if b != 0 {
			allZero = false
			break
		}
	}
	if allZero {
		return ""
	}
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		raw[0:4],
		raw[4:6],
		raw[6:8],
		raw[8:10],
		raw[10:16],
	)
}

func normalizeTransports(transports []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, transport := range transports {
		transport = strings.TrimSpace(strings.ToLower(transport))
		if transport == "" || seen[transport] {
			continue
		}
		seen[transport] = true
		out = append(out, transport)
	}
	return out
}
