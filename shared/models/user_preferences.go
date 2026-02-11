package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ============================================================================
// ENUMS
// ============================================================================

type ThemeMode string

const (
	ThemeModeLight ThemeMode = "light"
	ThemeModeDark  ThemeMode = "dark"
	ThemeModeAuto  ThemeMode = "auto"
)

// ============================================================================
// NOTIFICATION SETTINGS
// ============================================================================

type NotificationSettings struct {
	Email NotificationChannel `json:"email"`
	Push  NotificationChannel `json:"push"`
	InApp InAppNotifications  `json:"in_app"`
	SMS   NotificationChannel `json:"sms"`
}

type NotificationChannel struct {
	Enabled bool                    `json:"enabled"`
	Digest  string                  `json:"digest,omitempty"` // daily, weekly, instant
	Events  map[string]bool         `json:"events"`
}

type InAppNotifications struct {
	Enabled              bool `json:"enabled"`
	Sound                bool `json:"sound"`
	DesktopNotifications bool `json:"desktop_notifications"`
}

func (ns NotificationSettings) Value() (driver.Value, error) {
	return json.Marshal(ns)
}

func (ns *NotificationSettings) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal NotificationSettings")
	}
	return json.Unmarshal(bytes, ns)
}

// ============================================================================
// DASHBOARD SETTINGS
// ============================================================================

type DashboardSettings struct {
	Widgets     []DashboardWidget `json:"widgets"`
	DefaultView string            `json:"default_view"`
	ShowTips    bool              `json:"show_tips"`
}

type DashboardWidget struct {
	ID      string `json:"id"`
	Enabled bool   `json:"enabled"`
	Order   int    `json:"order"`
}

func (ds DashboardSettings) Value() (driver.Value, error) {
	return json.Marshal(ds)
}

func (ds *DashboardSettings) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal DashboardSettings")
	}
	return json.Unmarshal(bytes, ds)
}

// ============================================================================
// PRIVACY SETTINGS
// ============================================================================

type PrivacySettings struct {
	ShowProfilePicture bool `json:"show_profile_picture"`
	ShowPhone          bool `json:"show_phone"`
	ShowEmail          bool `json:"show_email"`
	ShowBirthday       bool `json:"show_birthday"`
	AllowMentions      bool `json:"allow_mentions"`
}

func (ps PrivacySettings) Value() (driver.Value, error) {
	return json.Marshal(ps)
}

func (ps *PrivacySettings) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal PrivacySettings")
	}
	return json.Unmarshal(bytes, ps)
}

// ============================================================================
// ACCESSIBILITY SETTINGS
// ============================================================================

type AccessibilitySettings struct {
	HighContrast       bool   `json:"high_contrast"`
	ReduceMotion       bool   `json:"reduce_motion"`
	ScreenReader       bool   `json:"screen_reader"`
	FontSize           string `json:"font_size"`
	KeyboardShortcuts  bool   `json:"keyboard_shortcuts"`
}

func (as AccessibilitySettings) Value() (driver.Value, error) {
	return json.Marshal(as)
}

func (as *AccessibilitySettings) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal AccessibilitySettings")
	}
	return json.Unmarshal(bytes, as)
}

// ============================================================================
// MOBILE SETTINGS
// ============================================================================

type MobileSettings struct {
	BiometricAuth bool `json:"biometric_auth"`
	OfflineMode   bool `json:"offline_mode"`
	DataSaver     bool `json:"data_saver"`
	AutoSync      bool `json:"auto_sync"`
}

func (ms MobileSettings) Value() (driver.Value, error) {
	return json.Marshal(ms)
}

func (ms *MobileSettings) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal MobileSettings")
	}
	return json.Unmarshal(bytes, ms)
}

// ============================================================================
// USER PREFERENCES MODEL
// ============================================================================

type UserPreferences struct {
	// Primary identification
	ID     uuid.UUID `db:"id" json:"id"`
	UserID uuid.UUID `db:"user_id" json:"user_id"`

	// Appearance
	Theme            ThemeMode `db:"theme" json:"theme"`
	CompactMode      bool      `db:"compact_mode" json:"compact_mode"`
	SidebarCollapsed bool      `db:"sidebar_collapsed" json:"sidebar_collapsed"`

	// Language & Locale
	Language   string `db:"language" json:"language"`
	Timezone   string `db:"timezone" json:"timezone"`
	DateFormat string `db:"date_format" json:"date_format"`
	TimeFormat string `db:"time_format" json:"time_format"`
	Currency   string `db:"currency" json:"currency"`

	// Notifications
	Notifications NotificationSettings `db:"notifications" json:"notifications"`

	// Dashboard
	Dashboard DashboardSettings `db:"dashboard" json:"dashboard"`

	// Privacy
	Privacy PrivacySettings `db:"privacy" json:"privacy"`

	// Accessibility
	Accessibility AccessibilitySettings `db:"accessibility" json:"accessibility"`

	// Email Digest
	EmailDigestFrequency string    `db:"email_digest_frequency" json:"email_digest_frequency"`
	EmailDigestTime      time.Time `db:"email_digest_time" json:"email_digest_time"`

	// Mobile
	MobileSettings MobileSettings `db:"mobile_settings" json:"mobile_settings"`

	// Quick Actions & Shortcuts
	QuickActions       json.RawMessage `db:"quick_actions" json:"quick_actions"`
	KeyboardShortcuts  json.RawMessage `db:"keyboard_shortcuts" json:"keyboard_shortcuts"`

	// Timestamps
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// ============================================================================
// HELPER METHODS
// ============================================================================

// IsNotificationEnabled checks if notification is enabled for event
func (up *UserPreferences) IsNotificationEnabled(channel, event string) bool {
	switch channel {
	case "email":
		return up.Notifications.Email.Enabled && up.Notifications.Email.Events[event]
	case "push":
		return up.Notifications.Push.Enabled && up.Notifications.Push.Events[event]
	case "sms":
		return up.Notifications.SMS.Enabled && up.Notifications.SMS.Events[event]
	case "in_app":
		return up.Notifications.InApp.Enabled
	default:
		return false
	}
}

// UpdateNotificationEvent updates notification setting for specific event
func (up *UserPreferences) UpdateNotificationEvent(channel, event string, enabled bool) {
	switch channel {
	case "email":
		if up.Notifications.Email.Events == nil {
			up.Notifications.Email.Events = make(map[string]bool)
		}
		up.Notifications.Email.Events[event] = enabled
	case "push":
		if up.Notifications.Push.Events == nil {
			up.Notifications.Push.Events = make(map[string]bool)
		}
		up.Notifications.Push.Events[event] = enabled
	case "sms":
		if up.Notifications.SMS.Events == nil {
			up.Notifications.SMS.Events = make(map[string]bool)
		}
		up.Notifications.SMS.Events[event] = enabled
	}
}