package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/phyowaiyan-dev/goappmon/internal/models"
	"github.com/phyowaiyan-dev/goappmon/internal/repositories"
	"github.com/phyowaiyan-dev/goappmon/internal/utils"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
)

var ErrPlatformDisabled = errors.New("platform disabled")
var ErrInvalidVersionFormat = errors.New("invalid version format")
var ErrLatestVersionNotGreater = errors.New("latest version must be greater than current version")
var ErrMinimumVersionGreaterThanLatest = errors.New("minimum version must be less than or equal to latest version")
var ErrCannotDeleteLastVersion = errors.New("cannot delete last version")

type ActionMeta struct {
	ActorID   int64
	IP        string
	UserAgent string
}

type AdminDashboard struct {
	Settings           *models.Setting
	Flags              []models.FeatureFlag
	AndroidReleases    []models.VersionRelease
	IOSReleases        []models.VersionRelease
	MaintenanceHistory []models.StateChange
	BannerHistory      []models.StateChange
	AuditLogs          []models.AuditLog
	SystemHealth       SystemHealth
}

type AuditLogSearchParams struct {
	Query    string
	Page     int
	PageSize int
}

type AuditLogSearchResult struct {
	Logs       []models.AuditLog
	Total      int64
	Page       int
	PageSize   int
	TotalPages int
}

type SystemHealth struct {
	Score             int
	Status            string
	ScoreTone         string
	StatusTone        string
	Uptime            string
	SystemUptime      string
	GoVersion         string
	NumCPU            int
	CPUUsagePercent   float64
	NumGoroutine      int
	AllocMB           float64
	TotalAllocMB      float64
	SysMB             float64
	MemoryTotalMB     float64
	MemoryUsedMB      float64
	MemoryFreeMB      float64
	MemoryUsedPercent float64
	DiskTotalGB       float64
	DiskUsedGB        float64
	DiskFreeGB        float64
	DiskUsedPercent   float64
	SQLiteFileSize    string
	SQLiteFilePath    string
	HealthNotes       []string
	UpdatedAt         time.Time
}

type AdminService struct {
	db        *sql.DB
	settings  *repositories.SettingRepository
	flags     *repositories.FeatureFlagRepository
	versions  *repositories.VersionReleaseRepository
	states    *repositories.StateHistoryRepository
	audits    *repositories.AuditLogRepository
	dbPath    string
	startedAt time.Time
}

func NewAdminService(db *sql.DB, settings *repositories.SettingRepository, flags *repositories.FeatureFlagRepository, dbPath string, startedAt time.Time) *AdminService {
	return &AdminService{
		db:        db,
		settings:  settings,
		flags:     flags,
		versions:  repositories.NewVersionReleaseRepository(db),
		states:    repositories.NewStateHistoryRepository(db),
		audits:    repositories.NewAuditLogRepository(db),
		dbPath:    dbPath,
		startedAt: startedAt,
	}
}

func (s *AdminService) Dashboard(ctx context.Context) (*AdminDashboard, error) {
	settings, err := s.settings.GetCurrent(ctx)
	if err != nil {
		return nil, err
	}
	flags, err := s.flags.List(ctx)
	if err != nil {
		return nil, err
	}
	androidReleases, err := s.versions.ListByPlatform(ctx, "android", 10)
	if err != nil {
		return nil, err
	}
	iosReleases, err := s.versions.ListByPlatform(ctx, "ios", 10)
	if err != nil {
		return nil, err
	}
	maintenanceHistory, err := s.states.ListByKind(ctx, "maintenance", 5)
	if err != nil {
		return nil, err
	}
	bannerHistory, err := s.states.ListByKind(ctx, "banner", 5)
	if err != nil {
		return nil, err
	}
	auditLogs, err := s.audits.ListRecent(ctx, 10)
	if err != nil {
		return nil, err
	}
	return &AdminDashboard{
		Settings:           settings,
		Flags:              flags,
		AndroidReleases:    androidReleases,
		IOSReleases:        iosReleases,
		MaintenanceHistory: maintenanceHistory,
		BannerHistory:      bannerHistory,
		AuditLogs:          auditLogs,
		SystemHealth:       s.SystemHealth(),
	}, nil
}

func (s *AdminService) SearchAuditLogs(ctx context.Context, params AuditLogSearchParams) (AuditLogSearchResult, error) {
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 20
	}
	filter := repositories.AuditLogFilter{
		Query:  params.Query,
		Limit:  params.PageSize,
		Offset: (params.Page - 1) * params.PageSize,
	}
	page, err := s.audits.Search(ctx, filter)
	if err != nil {
		return AuditLogSearchResult{}, err
	}
	totalPages := 0
	if page.Total > 0 {
		totalPages = int((page.Total + int64(params.PageSize) - 1) / int64(params.PageSize))
	}
	if totalPages == 0 {
		totalPages = 1
	}
	return AuditLogSearchResult{
		Logs:       page.Logs,
		Total:      page.Total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	}, nil
}

func (s *AdminService) UpdateApplication(ctx context.Context, meta ActionMeta, appName, apiURL string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	settings := repositories.NewSettingRepository(tx)
	audits := repositories.NewAuditLogRepository(tx)
	before, err := settings.GetCurrent(ctx)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := settings.UpdateApplication(ctx, appName, apiURL); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := createAudit(ctx, audits, meta, "settings.updated", "settings", "application", before, map[string]string{"app_name": appName, "api_url": apiURL}); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (s *AdminService) UpdatePlatforms(ctx context.Context, meta ActionMeta, androidEnabled, iosEnabled bool) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	settings := repositories.NewSettingRepository(tx)
	audits := repositories.NewAuditLogRepository(tx)
	before, err := settings.GetCurrent(ctx)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := settings.UpdatePlatforms(ctx, androidEnabled, iosEnabled); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := createAudit(ctx, audits, meta, "platforms.updated", "settings", "platforms", before, map[string]bool{"android_enabled": androidEnabled, "ios_enabled": iosEnabled}); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (s *AdminService) PublishVersion(ctx context.Context, meta ActionMeta, platform, latestVersion, minimumVersion string, forceUpdate bool, releaseNotes string) error {
	if _, err := utils.ParseSemanticVersion(latestVersion); err != nil {
		return ErrInvalidVersionFormat
	}
	if _, err := utils.ParseSemanticVersion(minimumVersion); err != nil {
		return ErrInvalidVersionFormat
	}
	if cmp, err := utils.CompareSemanticVersion(minimumVersion, latestVersion); err != nil {
		return ErrInvalidVersionFormat
	} else if cmp > 0 {
		return ErrMinimumVersionGreaterThanLatest
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	settings := repositories.NewSettingRepository(tx)
	versions := repositories.NewVersionReleaseRepository(tx)
	audits := repositories.NewAuditLogRepository(tx)
	before, err := settings.GetCurrent(ctx)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	switch platform {
	case "android":
		if !before.AndroidEnabled {
			_ = tx.Rollback()
			return ErrPlatformDisabled
		}
	case "ios":
		if !before.IOSEnabled {
			_ = tx.Rollback()
			return ErrPlatformDisabled
		}
	}
	if latest, err := versions.LatestByPlatform(ctx, platform); err == nil && latest != nil {
		if cmp, err := utils.CompareSemanticVersion(latestVersion, latest.LatestVersion); err != nil {
			_ = tx.Rollback()
			return ErrInvalidVersionFormat
		} else if cmp <= 0 {
			_ = tx.Rollback()
			return ErrLatestVersionNotGreater
		}
	}
	now := time.Now().UTC()
	release := models.VersionRelease{
		Platform:         platform,
		LatestVersion:    latestVersion,
		MinimumVersion:   minimumVersion,
		ForceUpdate:      forceUpdate,
		ReleaseNotes:     releaseNotes,
		CreatedByAdminID: meta.ActorID,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if _, err := versions.Create(ctx, release); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := settings.UpdatePlatformVersion(ctx, platform, latestVersion, minimumVersion, forceUpdate); err != nil {
		_ = tx.Rollback()
		return err
	}
	payload := map[string]any{
		"platform":        platform,
		"latest_version":  latestVersion,
		"minimum_version": minimumVersion,
		"force_update":    forceUpdate,
		"release_notes":   releaseNotes,
	}
	if err := createAudit(ctx, audits, meta, "version.created", "version_release", platform, before, payload); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (s *AdminService) DeleteCurrentVersion(ctx context.Context, meta ActionMeta, platform string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	settings := repositories.NewSettingRepository(tx)
	versions := repositories.NewVersionReleaseRepository(tx)
	audits := repositories.NewAuditLogRepository(tx)
	before, err := settings.GetCurrent(ctx)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	if platform != "android" && platform != "ios" {
		_ = tx.Rollback()
		return fmt.Errorf("invalid platform")
	}

	releases, err := versions.ListByPlatform(ctx, platform, 2)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	if len(releases) <= 1 {
		_ = tx.Rollback()
		return ErrCannotDeleteLastVersion
	}
	deleted := releases[0]
	promoted := releases[1]

	if err := versions.DeleteByID(ctx, deleted.ID); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := settings.UpdatePlatformVersion(ctx, platform, promoted.LatestVersion, promoted.MinimumVersion, promoted.ForceUpdate); err != nil {
		_ = tx.Rollback()
		return err
	}
	payload := map[string]any{
		"platform":         platform,
		"deleted_version":  deleted.LatestVersion,
		"promoted_version": promoted.LatestVersion,
	}
	if err := createAudit(ctx, audits, meta, "version.deleted", "version_release", platform, before, payload); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (s *AdminService) UpdateMaintenance(ctx context.Context, meta ActionMeta, enabled bool, message string) error {
	return s.updateState(ctx, meta, "maintenance", "maintenance.updated", enabled, message)
}

func (s *AdminService) UpdateBanner(ctx context.Context, meta ActionMeta, enabled bool, message string) error {
	return s.updateState(ctx, meta, "banner", "banner.updated", enabled, message)
}

func (s *AdminService) updateState(ctx context.Context, meta ActionMeta, kind, action string, enabled bool, message string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	settings := repositories.NewSettingRepository(tx)
	states := repositories.NewStateHistoryRepository(tx)
	audits := repositories.NewAuditLogRepository(tx)
	before, err := settings.GetCurrent(ctx)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	if !enabled {
		message = ""
	}
	now := time.Now().UTC()
	change := models.StateChange{
		Kind:             kind,
		Enabled:          enabled,
		Message:          message,
		CreatedByAdminID: meta.ActorID,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if _, err := states.Create(ctx, change); err != nil {
		_ = tx.Rollback()
		return err
	}
	switch kind {
	case "maintenance":
		if err := settings.UpdateMaintenance(ctx, enabled, message); err != nil {
			_ = tx.Rollback()
			return err
		}
	case "banner":
		if err := settings.UpdateBanner(ctx, enabled, message); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	payload := map[string]any{
		"enabled": enabled,
		"message": message,
		"kind":    kind,
	}
	if err := createAudit(ctx, audits, meta, action, "state_change", kind, before, payload); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (s *AdminService) CreateFlag(ctx context.Context, meta ActionMeta, key string, enabled bool) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	flags := repositories.NewFeatureFlagRepository(tx)
	audits := repositories.NewAuditLogRepository(tx)
	id, err := flags.Create(ctx, key, enabled)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	flag, err := flags.GetByID(ctx, id)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := createAudit(ctx, audits, meta, "flag.created", "feature_flag", fmt.Sprintf("%d", id), nil, flag); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (s *AdminService) UpdateFlag(ctx context.Context, meta ActionMeta, id int64, key string, enabled bool) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	flags := repositories.NewFeatureFlagRepository(tx)
	audits := repositories.NewAuditLogRepository(tx)
	before, err := flags.GetByID(ctx, id)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := flags.Update(ctx, id, key, enabled); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := createAudit(ctx, audits, meta, "flag.updated", "feature_flag", fmt.Sprintf("%d", id), before, map[string]any{"key": key, "enabled": enabled}); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (s *AdminService) DeleteFlag(ctx context.Context, meta ActionMeta, id int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	flags := repositories.NewFeatureFlagRepository(tx)
	audits := repositories.NewAuditLogRepository(tx)
	before, err := flags.GetByID(ctx, id)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := flags.Delete(ctx, id); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := createAudit(ctx, audits, meta, "flag.deleted", "feature_flag", fmt.Sprintf("%d", id), before, nil); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (s *AdminService) SystemHealth() SystemHealth {
	var runtimeMem runtime.MemStats
	runtime.ReadMemStats(&runtimeMem)

	ctx, cancel := context.WithTimeout(context.Background(), 750*time.Millisecond)
	defer cancel()

	dbSize := s.sqliteFileSize()
	notes := make([]string, 0, 6)

	memInfo, memErr := mem.VirtualMemoryWithContext(ctx)
	diskInfo, diskErr := disk.UsageWithContext(ctx, s.diskUsagePath())
	cpuUsage := 0.0
	if cpuPercents, err := cpu.PercentWithContext(ctx, 200*time.Millisecond, false); err == nil && len(cpuPercents) > 0 {
		cpuUsage = cpuPercents[0]
	}
	systemUptime := ""
	if uptimeSeconds, err := host.UptimeWithContext(ctx); err == nil && uptimeSeconds > 0 {
		systemUptime = (time.Duration(uptimeSeconds) * time.Second).Truncate(time.Second).String()
	}
	if memErr != nil {
		notes = append(notes, "System memory snapshot is unavailable")
	}
	if diskErr != nil {
		notes = append(notes, "Disk usage snapshot is unavailable")
	}

	score, status, scoreNotes := calculateHealthScore(cpuUsage, memInfo, memErr, diskInfo, diskErr, runtime.NumGoroutine(), dbSize, runtimeMem.Alloc)
	notes = append(notes, scoreNotes...)

	return SystemHealth{
		Score:           score,
		Status:          status,
		ScoreTone:       healthTone(status),
		StatusTone:      healthTone(status),
		Uptime:          time.Since(s.startedAt).Truncate(time.Second).String(),
		SystemUptime:    systemUptime,
		GoVersion:       runtime.Version(),
		NumCPU:          runtime.NumCPU(),
		CPUUsagePercent: cpuUsage,
		NumGoroutine:    runtime.NumGoroutine(),
		AllocMB:         bytesToMB(runtimeMem.Alloc),
		TotalAllocMB:    bytesToMB(runtimeMem.TotalAlloc),
		SysMB:           bytesToMB(runtimeMem.Sys),
		MemoryTotalMB: func() float64 {
			if memErr != nil {
				return 0
			}
			return bytesToMB(memInfo.Total)
		}(),
		MemoryUsedMB: func() float64 {
			if memErr != nil {
				return 0
			}
			return bytesToMB(memInfo.Used)
		}(),
		MemoryFreeMB: func() float64 {
			if memErr != nil {
				return 0
			}
			return bytesToMB(memInfo.Available)
		}(),
		MemoryUsedPercent: func() float64 {
			if memErr != nil {
				return 0
			}
			return memInfo.UsedPercent
		}(),
		DiskTotalGB: func() float64 {
			if diskErr != nil {
				return 0
			}
			return bytesToGB(diskInfo.Total)
		}(),
		DiskUsedGB: func() float64 {
			if diskErr != nil {
				return 0
			}
			return bytesToGB(diskInfo.Used)
		}(),
		DiskFreeGB: func() float64 {
			if diskErr != nil {
				return 0
			}
			return bytesToGB(diskInfo.Free)
		}(),
		DiskUsedPercent: func() float64 {
			if diskErr != nil {
				return 0
			}
			return diskInfo.UsedPercent
		}(),
		SQLiteFileSize: humanFileSize(dbSize),
		SQLiteFilePath: s.dbPath,
		HealthNotes:    notes,
		UpdatedAt:      time.Now().UTC(),
	}
}

func calculateHealthScore(cpuUsage float64, memInfo *mem.VirtualMemoryStat, memErr error, diskInfo *disk.UsageStat, diskErr error, goroutines int, dbSize int64, appAllocBytes uint64) (int, string, []string) {
	score := 96.0
	notes := make([]string, 0, 6)

	if memErr == nil && memInfo != nil {
		switch {
		case memInfo.UsedPercent > 90:
			score -= (memInfo.UsedPercent - 70) * 1.1
			notes = append(notes, "Memory is in a critical zone")
		case memInfo.UsedPercent > 70:
			score -= (memInfo.UsedPercent - 70) * 0.9
		}
		if memInfo.Available < 1024*1024*1024 {
			score -= 6
			notes = append(notes, "Available memory is below 1 GB")
		}
	}

	if diskErr == nil && diskInfo != nil {
		switch {
		case diskInfo.UsedPercent > 95:
			score -= (diskInfo.UsedPercent - 75) * 1.25
			notes = append(notes, "Disk usage is critically high")
		case diskInfo.UsedPercent > 75:
			score -= (diskInfo.UsedPercent - 75) * 1.25
		}
		if diskInfo.Free < 10*1024*1024*1024 {
			score -= 6
			notes = append(notes, "Disk free space is below 10 GB")
		}
	}

	if cpuUsage > 80 {
		score -= (cpuUsage - 80) * 0.35
		notes = append(notes, "CPU usage is elevated")
	} else if cpuUsage > 55 {
		score -= (cpuUsage - 55) * 0.22
	}

	if goroutines > 75 {
		score -= 2
	}
	if dbSize > 100*1024*1024 {
		score -= 4
	}
	if appAllocBytes > 256*1024*1024 {
		score -= 4
	}

	if score > 96 {
		score = 96
	}
	if score < 36 {
		score = 36
	}

	status := "Healthy"
	switch {
	case score >= 90:
		status = "Excellent"
	case score >= 75:
		status = "Good"
	case score >= 55:
		status = "Fair"
	default:
		status = "Needs attention"
	}

	return int(score + 0.5), status, notes
}

func (s *AdminService) sqliteFileSize() int64 {
	if s.dbPath == "" {
		return 0
	}
	info, err := os.Stat(s.dbPath)
	if err != nil {
		return 0
	}
	return info.Size()
}

func (s *AdminService) diskUsagePath() string {
	if s.dbPath == "" {
		return "."
	}
	dir := filepath.Dir(s.dbPath)
	if dir == "" || dir == "." {
		return s.dbPath
	}
	return dir
}

func createAudit(ctx context.Context, repo *repositories.AuditLogRepository, meta ActionMeta, action, entityType, entityID string, before, after any) error {
	entry := models.AuditLog{
		ActorAdminID: meta.ActorID,
		Action:       action,
		EntityType:   entityType,
		EntityID:     entityID,
		IP:           meta.IP,
		UserAgent:    meta.UserAgent,
		CreatedAt:    time.Now().UTC(),
	}
	if before != nil {
		data, err := json.Marshal(before)
		if err != nil {
			return err
		}
		entry.BeforeJSON = string(data)
	}
	if after != nil {
		data, err := json.Marshal(after)
		if err != nil {
			return err
		}
		entry.AfterJSON = string(data)
	}
	_, err := repo.Create(ctx, entry)
	return err
}

func bytesToMB(v uint64) float64 {
	return float64(v) / (1024 * 1024)
}

func bytesToGB(v uint64) float64 {
	return float64(v) / (1024 * 1024 * 1024)
}

func humanFileSize(bytes int64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	}
	unit := []string{"KB", "MB", "GB", "TB"}
	size := float64(bytes)
	i := 0
	for size >= 1024 && i < len(unit)-1 {
		size /= 1024
		i++
	}
	return fmt.Sprintf("%.1f %s", size, unit[i])
}

func healthTone(status string) string {
	switch strings.ToLower(status) {
	case "healthy":
		return "success"
	case "excellent":
		return "success"
	case "good":
		return "info"
	case "fair":
		return "warning"
	default:
		return "error"
	}
}
