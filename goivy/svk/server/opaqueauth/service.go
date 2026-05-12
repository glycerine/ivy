package opaqueauth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"

	"github.com/bytemare/opaque"
)

type Service struct {
	mu sync.Mutex

	conf             *opaque.Configuration
	serverID         []byte
	serverPrivateKey []byte
	serverPublicKey  []byte
	oprfSeed         []byte

	registrationFlows map[string]string
	loginFlows        map[string]*opaque.Server
	records           map[string]*opaque.ClientRecord
}

func NewService(serverID []byte) (*Service, error) {
	conf := opaque.DefaultConfiguration()
	secret, public := conf.KeyGen()
	if len(secret) == 0 || len(public) == 0 {
		return nil, errors.New("opaque key generation failed")
	}
	return &Service{
		conf:              conf,
		serverID:          append([]byte(nil), serverID...),
		serverPrivateKey:  secret,
		serverPublicKey:   public,
		oprfSeed:          conf.GenerateOPRFSeed(),
		registrationFlows: map[string]string{},
		loginFlows:        map[string]*opaque.Server{},
		records:           map[string]*opaque.ClientRecord{},
	}, nil
}

func (s *Service) Configuration() []byte {
	return s.conf.Serialize()
}

func (s *Service) ServerPublicKey() []byte {
	return append([]byte(nil), s.serverPublicKey...)
}

func (s *Service) ServerIdentity() []byte {
	return append([]byte(nil), s.serverID...)
}

func (s *Service) RegistrationStart(userID string, requestBytes []byte) (flowID string, responseBytes []byte, err error) {
	server, err := s.conf.Server()
	if err != nil {
		return "", nil, err
	}
	request, err := server.Deserialize.RegistrationRequest(requestBytes)
	if err != nil {
		return "", nil, err
	}
	publicKey, err := server.Deserialize.DecodeAkePublicKey(s.serverPublicKey)
	if err != nil {
		return "", nil, err
	}
	response := server.RegistrationResponse(request, publicKey, []byte(userID), s.oprfSeed)
	flowID = randomID()

	s.mu.Lock()
	s.registrationFlows[flowID] = userID
	s.mu.Unlock()

	return flowID, response.Serialize(), nil
}

func (s *Service) RegistrationFinish(flowID string, recordBytes []byte) error {
	s.mu.Lock()
	userID, ok := s.registrationFlows[flowID]
	if ok {
		delete(s.registrationFlows, flowID)
	}
	s.mu.Unlock()
	if !ok {
		return errors.New("unknown opaque registration flow")
	}

	server, err := s.conf.Server()
	if err != nil {
		return err
	}
	record, err := server.Deserialize.RegistrationRecord(recordBytes)
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.records[userID] = &opaque.ClientRecord{
		CredentialIdentifier: []byte(userID),
		ClientIdentity:       []byte(userID),
		RegistrationRecord:   record,
	}
	s.mu.Unlock()
	return nil
}

func (s *Service) LoginStart(userID string, ke1Bytes []byte) (flowID string, ke2Bytes []byte, err error) {
	server, err := s.conf.Server()
	if err != nil {
		return "", nil, err
	}
	if err := server.SetKeyMaterial(s.serverID, s.serverPrivateKey, s.serverPublicKey, s.oprfSeed); err != nil {
		return "", nil, err
	}
	ke1, err := server.Deserialize.KE1(ke1Bytes)
	if err != nil {
		return "", nil, err
	}

	s.mu.Lock()
	record := s.records[userID]
	s.mu.Unlock()
	if record == nil {
		record, err = s.conf.GetFakeRecord([]byte(userID))
		if err != nil {
			return "", nil, err
		}
	}
	ke2, err := server.LoginInit(ke1, record)
	if err != nil {
		return "", nil, err
	}
	flowID = randomID()
	s.mu.Lock()
	s.loginFlows[flowID] = server
	s.mu.Unlock()
	return flowID, ke2.Serialize(), nil
}

func (s *Service) LoginFinish(flowID string, ke3Bytes []byte) ([]byte, error) {
	s.mu.Lock()
	server, ok := s.loginFlows[flowID]
	if ok {
		delete(s.loginFlows, flowID)
	}
	s.mu.Unlock()
	if !ok {
		return nil, errors.New("unknown opaque login flow")
	}
	ke3, err := server.Deserialize.KE3(ke3Bytes)
	if err != nil {
		return nil, err
	}
	if err := server.LoginFinish(ke3); err != nil {
		return nil, err
	}
	return server.SessionKey(), nil
}

func randomID() string {
	var buf [16]byte
	if _, err := rand.Read(buf[:]); err != nil {
		panic(err)
	}
	return hex.EncodeToString(buf[:])
}
