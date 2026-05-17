package server

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bytemare/opaque"
)

func TestOpaqueRoutesCompleteRegistrationAndLoginWithoutRawPassword(t *testing.T) {
	srv, err := New(Config{DevMode: true})
	if err != nil {
		t.Fatal(err)
	}
	conf, err := opaque.DeserializeConfiguration(srv.opaque.Configuration())
	if err != nil {
		t.Fatal(err)
	}
	password := []byte("correct horse battery staple")
	userID := "user-opaque-1"

	client, err := conf.Client()
	if err != nil {
		t.Fatal(err)
	}
	reg1 := client.RegistrationInit(password)
	regStart := postOpaque(t, srv, "/auth/opaque/register/start", map[string]string{
		"userId":  userID,
		"message": base64.StdEncoding.EncodeToString(reg1.Serialize()),
	})
	regRespBytes, err := base64.StdEncoding.DecodeString(regStart.Message)
	if err != nil {
		t.Fatal(err)
	}
	regResp, err := client.Deserialize.RegistrationResponse(regRespBytes)
	if err != nil {
		t.Fatal(err)
	}
	record, _ := client.RegistrationFinalize(regResp, opaque.ClientRegistrationFinalizeOptions{
		ClientIdentity: []byte(userID),
		ServerIdentity: srv.opaque.ServerIdentity(),
	})
	regFinish := postOpaque(t, srv, "/auth/opaque/register/finish", map[string]string{
		"flowId":  regStart.FlowID,
		"message": base64.StdEncoding.EncodeToString(record.Serialize()),
	})
	if !regFinish.Complete {
		t.Fatal("registration did not complete")
	}

	loginClient, err := conf.Client()
	if err != nil {
		t.Fatal(err)
	}
	ke1 := loginClient.LoginInit(password)
	loginStart := postOpaque(t, srv, "/auth/opaque/login/start", map[string]string{
		"userId":  userID,
		"message": base64.StdEncoding.EncodeToString(ke1.Serialize()),
	})
	ke2Bytes, err := base64.StdEncoding.DecodeString(loginStart.Message)
	if err != nil {
		t.Fatal(err)
	}
	ke2, err := loginClient.Deserialize.KE2(ke2Bytes)
	if err != nil {
		t.Fatal(err)
	}
	ke3, _, err := loginClient.LoginFinish(ke2, opaque.ClientLoginFinishOptions{
		ClientIdentity: []byte(userID),
		ServerIdentity: srv.opaque.ServerIdentity(),
	})
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	body := encodeJSON(t, map[string]string{
		"flowId":  loginStart.FlowID,
		"message": base64.StdEncoding.EncodeToString(ke3.Serialize()),
	})
	req := httptest.NewRequest(http.MethodPost, "/auth/opaque/login/finish", body)
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("login finish status = %d body=%s", rec.Code, rec.Body.String())
	}
	if len(rec.Result().Cookies()) == 0 {
		t.Fatal("expected login finish to set a session cookie")
	}
}

func TestOpaqueRoutesRejectRawPasswordFields(t *testing.T) {
	srv, err := New(Config{DevMode: true})
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/auth/opaque/register/start", encodeJSON(t, map[string]string{
		"userId":   "user-1",
		"message":  "not-base64",
		"password": "never send this",
	}))
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("raw passwords are not accepted")) {
		t.Fatalf("body = %s", rec.Body.String())
	}
}

func postOpaque(t *testing.T, handler http.Handler, path string, value map[string]string) opaqueMessageResponse {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, path, encodeJSON(t, value))
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("%s status = %d body=%s", path, rec.Code, rec.Body.String())
	}
	var response opaqueMessageResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}
	return response
}

func encodeJSON(t *testing.T, value interface{}) *bytes.Reader {
	t.Helper()
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(value); err != nil {
		t.Fatal(err)
	}
	return bytes.NewReader(buf.Bytes())
}
