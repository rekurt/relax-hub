package domain

import "time"

// SettingType defines the data type of a platform setting value.
type SettingType string

const (
	SettingTypeInt    SettingType = "int"
	SettingTypeFloat  SettingType = "float"
	SettingTypeString SettingType = "string"
	SettingTypeBool   SettingType = "bool"
	SettingTypeJSON   SettingType = "json"
)

func (t SettingType) IsValid() bool {
	switch t {
	case SettingTypeInt, SettingTypeFloat, SettingTypeString, SettingTypeBool, SettingTypeJSON:
		return true
	}
	return false
}

// PlatformSetting represents a single admin-configurable platform setting.
type PlatformSetting struct {
	Key         string      `json:"key"`
	Value       string      `json:"value"`
	Description string      `json:"description"`
	Type        SettingType `json:"type"`
	UpdatedAt   time.Time   `json:"updated_at"`
	UpdatedBy   *string     `json:"updated_by,omitempty"`
}

func (s *PlatformSetting) Validate() error {
	if s.Key == "" {
		return ErrInvalidInput
	}
	if !s.Type.IsValid() {
		return ErrInvalidInput
	}
	return nil
}
