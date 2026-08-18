package setting

import "time"

// CoreConfig represents a system configuration key-value record matching Magento 2 core_config_data architecture.
type CoreConfig struct {
	ID        string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Scope     string    `json:"scope" gorm:"type:varchar(20);default:'default';not null;uniqueIndex:idx_scope_scope_id_path,priority:1"`
	ScopeID   string    `json:"scope_id" gorm:"type:varchar(36);default:'0';not null;uniqueIndex:idx_scope_scope_id_path,priority:2"`
	Path      string    `json:"path" gorm:"type:varchar(255);not null;uniqueIndex:idx_scope_scope_id_path,priority:3;index:idx_config_path"`
	Value     string    `json:"value" gorm:"type:text"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName overrides the default table name to match Magento 2's core_config_data convention.
func (CoreConfig) TableName() string {
	return "core_config_data"
}
