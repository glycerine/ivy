package control

import "time"

type User struct {
	ID              string     `json:"id"`
	IDPIssuer       string     `json:"idpIssuer"`
	IDPSubject      string     `json:"idpSubject"`
	Email           string     `json:"email"`
	DisplayName     string     `json:"displayName"`
	EmailVerifiedAt *time.Time `json:"emailVerifiedAt,omitempty"`
	DisabledAt      *time.Time `json:"disabledAt,omitempty"`
}

type Account struct {
	ID            string     `json:"id"`
	Slug          string     `json:"slug"`
	DisplayName   string     `json:"displayName"`
	BillingEmail  string     `json:"billingEmail"`
	BillingStatus string     `json:"billingStatus"`
	DisabledAt    *time.Time `json:"disabledAt,omitempty"`
}

type Team struct {
	ID          string     `json:"id"`
	AccountID   string     `json:"accountId"`
	Slug        string     `json:"slug"`
	DisplayName string     `json:"displayName"`
	DisabledAt  *time.Time `json:"disabledAt,omitempty"`
}

type Project struct {
	ID              string     `json:"id"`
	AccountID       string     `json:"accountId"`
	Slug            string     `json:"slug"`
	DisplayName     string     `json:"displayName"`
	CreatedByUserID string     `json:"createdByUserId"`
	DisabledAt      *time.Time `json:"disabledAt,omitempty"`
}

type ProjectRole string

const (
	ProjectRoleRead  ProjectRole = "read"
	ProjectRoleWrite ProjectRole = "write"
	ProjectRoleAdmin ProjectRole = "admin"
)

type ProjectGrantSubjectKind string

const (
	ProjectGrantSubjectUser    ProjectGrantSubjectKind = "user"
	ProjectGrantSubjectTeam    ProjectGrantSubjectKind = "team"
	ProjectGrantSubjectAccount ProjectGrantSubjectKind = "account"
)

type SessionView struct {
	Authenticated bool              `json:"authenticated"`
	User          *User             `json:"user,omitempty"`
	Accounts      []Account         `json:"accounts,omitempty"`
	Teams         []Team            `json:"teams,omitempty"`
	Projects      []Project         `json:"projects,omitempty"`
	Roles         map[string]string `json:"roles,omitempty"`
	Passkey       *PasskeyState     `json:"passkey,omitempty"`

	CookieRefreshNeeded bool `json:"-"`
}

type PasskeyState struct {
	Registered bool `json:"registered"`
}

type SessionTouch struct {
	UserID               string
	CookieRefreshNeeded  bool
	VisitHourInserted    bool
	PreviousLastSeenAt   time.Time
	CurrentVisitRecorded time.Time
}

type StoredPasskeyCredential struct {
	ID              string
	UserID          string
	CredentialID    []byte
	PublicKeyCOSE   []byte
	SignCount       uint32
	Transports      []string
	BackupEligible  bool
	BackedUp        bool
	AttestationType string
	AAGUID          string
	DisplayName     string
}

type BillingProvider interface {
	CreateCustomer(account Account, billingEmail string) (string, error)
	CreateSetupIntent(account Account) (string, error)
	AttachPaymentMethod(account Account, paymentMethodToken string) (string, error)
	MarkDefaultPaymentMethod(account Account, paymentMethodID string) error
}

type AnalysisClient interface {
	CreateSession(projectID, userID string) (string, error)
}
