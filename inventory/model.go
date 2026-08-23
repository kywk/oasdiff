package inventory

// ChangeType represents the type of change for an inventory record.
type ChangeType string

const (
	ChangeTypeNew    ChangeType = "新增"
	ChangeTypeUpdate ChangeType = "修改"
	ChangeTypeDelete ChangeType = "刪除"
)

// Record represents a single row in the inventory sheet.
type Record struct {
	Serial     string            `json:"serial"`      // e.g. "API-001"
	Date       string            `json:"date"`        // YYYY-MM-DD
	ChangeType ChangeType        `json:"change_type"` // 新增 / 修改 / 刪除
	Method     string            `json:"method"`      // HTTP method (identity key part 1)
	Path       string            `json:"path"`        // API path (identity key part 2)
	Values     map[string]string `json:"values"`      // column source -> value
}

// IdentityKey returns the unique identifier for this API endpoint.
func (r *Record) IdentityKey() string {
	return r.Method + " " + r.Path
}

// Sheet represents the full inventory spreadsheet.
type Sheet struct {
	Config    *Config  `json:"-"`
	Records   []Record `json:"records"`
	MaxSerial int      `json:"max_serial"` // current max serial number (including deleted)
}

// NewSheet creates an empty inventory sheet with the given config.
func NewSheet(cfg *Config) *Sheet {
	return &Sheet{
		Config:    cfg,
		Records:   make([]Record, 0),
		MaxSerial: 0,
	}
}

// ActiveRecords returns all records that are not deleted.
func (s *Sheet) ActiveRecords() []Record {
	result := make([]Record, 0)
	for _, r := range s.Records {
		if r.ChangeType != ChangeTypeDelete {
			result = append(result, r)
		}
	}
	return result
}

// FindByKey looks up a record by its identity key (method + path).
// Returns the record and its index, or nil and -1 if not found.
func (s *Sheet) FindByKey(method, path string) (*Record, int) {
	key := method + " " + path
	for i := range s.Records {
		if s.Records[i].IdentityKey() == key {
			return &s.Records[i], i
		}
	}
	return nil, -1
}

// FindActiveByKey looks up an active (non-deleted) record by its identity key.
func (s *Sheet) FindActiveByKey(method, path string) (*Record, int) {
	key := method + " " + path
	for i := range s.Records {
		if s.Records[i].IdentityKey() == key && s.Records[i].ChangeType != ChangeTypeDelete {
			return &s.Records[i], i
		}
	}
	return nil, -1
}
