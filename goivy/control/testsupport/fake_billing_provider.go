package testsupport

import (
	"fmt"
	"sync"

	"github.com/glycerine/ivy/goivy/control"
)

type BillingCall struct {
	Name      string
	AccountID string
	Value     string
}

type FakeBillingProvider struct {
	mu    sync.Mutex
	Calls []BillingCall
}

func (p *FakeBillingProvider) CreateCustomer(account control.Account, billingEmail string) (string, error) {
	p.record("CreateCustomer", account.ID, billingEmail)
	return fmt.Sprintf("cus_%s", account.ID), nil
}

func (p *FakeBillingProvider) CreateSetupIntent(account control.Account) (string, error) {
	p.record("CreateSetupIntent", account.ID, "")
	return fmt.Sprintf("seti_%s", account.ID), nil
}

func (p *FakeBillingProvider) AttachPaymentMethod(account control.Account, paymentMethodToken string) (string, error) {
	p.record("AttachPaymentMethod", account.ID, paymentMethodToken)
	return fmt.Sprintf("pm_%s", account.ID), nil
}

func (p *FakeBillingProvider) MarkDefaultPaymentMethod(account control.Account, paymentMethodID string) error {
	p.record("MarkDefaultPaymentMethod", account.ID, paymentMethodID)
	return nil
}

func (p *FakeBillingProvider) record(name, accountID, value string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Calls = append(p.Calls, BillingCall{Name: name, AccountID: accountID, Value: value})
}
