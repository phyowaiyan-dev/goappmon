package services

import (
	"context"

	"github.com/phyowaiyan-dev/goappmon/internal/models"
	"github.com/phyowaiyan-dev/goappmon/internal/repositories"
)

type StatusService struct {
	settings *repositories.SettingRepository
	flags    *repositories.FeatureFlagRepository
}

type PublicStatus struct {
	MaintenanceMode    bool   `json:"maintenance_mode"`
	MaintenanceMessage string `json:"maintenance_message"`
	BannerEnabled      bool   `json:"banner_enabled"`
	BannerMessage      string `json:"banner_message"`
}

type PublicVersion struct {
	Android struct {
		LatestVersion  string `json:"latest_version"`
		MinimumVersion string `json:"minimum_version"`
		ForceUpdate    bool   `json:"force_update"`
	} `json:"android"`
	IOS struct {
		LatestVersion  string `json:"latest_version"`
		MinimumVersion string `json:"minimum_version"`
		ForceUpdate    bool   `json:"force_update"`
	} `json:"ios"`
}

type PublicConfig struct {
	AppName string `json:"app_name"`
	APIURL  string `json:"api_url"`
}

func NewStatusService(settings *repositories.SettingRepository, flags *repositories.FeatureFlagRepository) *StatusService {
	return &StatusService{settings: settings, flags: flags}
}

func (s *StatusService) CurrentSettings(ctx context.Context) (*models.Setting, error) {
	return s.settings.GetCurrent(ctx)
}

func (s *StatusService) PublicStatus(ctx context.Context) (PublicStatus, error) {
	setting, err := s.settings.GetCurrent(ctx)
	if err != nil {
		return PublicStatus{}, err
	}
	return PublicStatus{
		MaintenanceMode:    setting.MaintenanceMode,
		MaintenanceMessage: setting.MaintenanceMessage,
		BannerEnabled:      setting.BannerEnabled,
		BannerMessage:      setting.BannerMessage,
	}, nil
}

func (s *StatusService) PublicVersion(ctx context.Context) (PublicVersion, error) {
	setting, err := s.settings.GetCurrent(ctx)
	if err != nil {
		return PublicVersion{}, err
	}
	var version PublicVersion
	version.Android.LatestVersion = setting.AndroidLatestVersion
	version.Android.MinimumVersion = setting.AndroidMinVersion
	version.Android.ForceUpdate = setting.AndroidForceUpdate
	version.IOS.LatestVersion = setting.IOSLatestVersion
	version.IOS.MinimumVersion = setting.IOSMinVersion
	version.IOS.ForceUpdate = setting.IOSForceUpdate
	return version, nil
}

func (s *StatusService) PublicConfig(ctx context.Context) (PublicConfig, error) {
	setting, err := s.settings.GetCurrent(ctx)
	if err != nil {
		return PublicConfig{}, err
	}
	return PublicConfig{
		AppName: setting.AppName,
		APIURL:  setting.APIURL,
	}, nil
}

func (s *StatusService) PublicFeatureFlags(ctx context.Context) (map[string]bool, error) {
	return s.flags.AsMap(ctx)
}
