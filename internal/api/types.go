package api

type Blueprint struct {
	ID          string     `json:"ID"`
	Slug        string     `json:"Slug"`
	Name        string     `json:"Name"`
	Description string     `json:"Description"`
	IsDefault   bool       `json:"IsDefault"`
	Fields      []FieldDef `json:"Fields"`
	CreatedAt   string     `json:"CreatedAt"`
}

type FieldDef struct {
	ID           string `json:"ID"`
	Name         string `json:"Name"`
	FieldType    string `json:"FieldType"`
	Required     bool   `json:"Required"`
	DefaultValue string `json:"DefaultValue"`
	SortOrder    int    `json:"SortOrder"`
}

type Entity struct {
	ID              string       `json:"ID"`
	Slug            string       `json:"Slug"`
	Name            string       `json:"Name"`
	BlueprintID     string       `json:"BlueprintID"`
	CurrentStatus   string       `json:"CurrentStatus"`
	StatusChangedAt string       `json:"StatusChangedAt"`
	Fields          []FieldValue `json:"Fields"`
	CreatedAt       string       `json:"CreatedAt"`
}

type FieldValue struct {
	ID         string `json:"ID"`
	FieldDefID string `json:"FieldDefID"`
	Value      string `json:"Value"`
}

type Check struct {
	ID               string                 `json:"ID"`
	EntityID         string                 `json:"EntityID"`
	CheckType        string                 `json:"CheckType"`
	TargetURL        string                 `json:"TargetURL"`
	IntervalSeconds  int                    `json:"IntervalSeconds"`
	TimeoutMs        int                    `json:"TimeoutMs"`
	ExpectedStatus   int                    `json:"ExpectedStatus"`
	BodyContains     string                 `json:"BodyContains"`
	FailureThreshold int                    `json:"FailureThreshold"`
	Enabled          bool                   `json:"Enabled"`
	NextRunAt        string                 `json:"NextRunAt"`
	Config           map[string]interface{} `json:"Config"`
	CreatedAt        string                 `json:"CreatedAt"`
}

type CheckResult struct {
	ID           string `json:"ID"`
	CheckID      string `json:"CheckID"`
	StatusCode   int    `json:"StatusCode"`
	LatencyMs    int    `json:"LatencyMs"`
	Success      bool   `json:"Success"`
	ErrorMessage string `json:"ErrorMessage"`
	CheckedAt    string `json:"CheckedAt"`
}

type CheckResultsResponse struct {
	Results    []CheckResult `json:"results"`
	NextCursor string        `json:"next_cursor"`
}

type StateTransition struct {
	ID              string `json:"ID"`
	EntityID        string `json:"EntityID"`
	CheckID         string `json:"CheckID"`
	FromState       string `json:"FromState"`
	ToState         string `json:"ToState"`
	TriggerReason   string `json:"TriggerReason"`
	TransitionedAt  string `json:"TransitionedAt"`
}

type StateTransitionWithEntity struct {
	StateTransition
	EntityName string `json:"EntityName"`
}

type NotifyPolicy struct {
	ID                    string `json:"ID"`
	CheckID               string `json:"CheckID"`
	CooldownSeconds       int    `json:"CooldownSeconds"`
	NotifyOnRecovery      bool   `json:"NotifyOnRecovery"`
	Silenced              bool   `json:"Silenced"`
	RepeatIntervalSeconds int    `json:"RepeatIntervalSeconds"`
	CreatedAt             string `json:"CreatedAt"`
}

type NotifyChannel struct {
	ID          string `json:"ID"`
	PolicyID    string `json:"PolicyID"`
	ChannelType string `json:"ChannelType"`
	ConfigJSON  string `json:"ConfigJSON"`
	Active      bool   `json:"Active"`
	CreatedAt   string `json:"CreatedAt"`
}

type NotifyLog struct {
	ID           string `json:"ID"`
	ChannelID    string `json:"ChannelID"`
	TransitionID string `json:"TransitionID"`
	Status       string `json:"Status"`
	Attempts     int    `json:"Attempts"`
	LastError    string `json:"LastError"`
	SentAt       string `json:"SentAt"`
}

type PolicyWithChannels struct {
	Policy   NotifyPolicy    `json:"policy"`
	Channels []NotifyChannel `json:"channels"`
}

type StatusEntity struct {
	ID            string `json:"id"`
	Slug          string `json:"slug"`
	Name          string `json:"name"`
	CurrentStatus string `json:"current_status"`
	BlueprintSlug string `json:"blueprint_slug"`
}

type StatusResponse struct {
	Entities []StatusEntity `json:"entities"`
	Total    int            `json:"total"`
	Up       int            `json:"up"`
	Degraded int            `json:"degraded"`
	Down     int            `json:"down"`
}

type EntityRelation struct {
	ID           string `json:"ID"`
	FromEntityID string `json:"FromEntityID"`
	ToEntityID   string `json:"ToEntityID"`
	RelationType string `json:"RelationType"`
	CreatedAt    string `json:"CreatedAt"`
}

type APIKey struct {
	ID         string `json:"id"`
	Prefix     string `json:"prefix"`
	Label      string `json:"label"`
	Active     bool   `json:"active"`
	LastUsedAt string `json:"last_used_at"`
	CreatedAt  string `json:"created_at"`
	ExpiresAt  string `json:"expires_at"`
}

type TeamUser struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	Role          string `json:"role"`
	EmailVerified bool   `json:"email_verified"`
	CreatedAt     string `json:"created_at"`
}

type TeamInvitation struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	ExpiresAt string `json:"expires_at"`
	CreatedAt string `json:"created_at"`
}

type TeamResponse struct {
	Users       []TeamUser       `json:"users"`
	Invitations []TeamInvitation `json:"invitations"`
}

type Skill struct {
	ID          string      `json:"ID"`
	Slug        string      `json:"Slug"`
	Name        string      `json:"Name"`
	Description string      `json:"Description"`
	Version     string      `json:"Version"`
	Tools       []SkillTool `json:"Tools"`
	Enabled     bool        `json:"Enabled"`
	CreatedAt   string      `json:"CreatedAt"`
	UpdatedAt   string      `json:"UpdatedAt"`
}

type SkillTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}
