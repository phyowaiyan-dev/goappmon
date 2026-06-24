package services

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/phyowaiyan-dev/goappmon/internal/models"
	"github.com/phyowaiyan-dev/goappmon/internal/repositories"
)

type AdminDashboard struct {
	Settings     *models.Setting
	Flags        []models.FeatureFlag
	SystemHealth SystemHealth
}

type SystemHealth struct {
	Score          int
	Status         string
	Uptime         string
	GoVersion      string
	NumCPU         int
	NumGoroutine   int
	AllocMB        float64
	TotalAllocMB   float64
	SysMB          float64
	SQLiteFileSize string
	SQLiteFilePath string
	HealthNotes    []string
}

type AdminService struct {
	settings  *repositories.SettingRepository
	flags     *repositories.FeatureFlagRepository
	dbPath    string
	startedAt time.Time
}

func NewAdminService(settings *repositories.SettingRepository, flags *repositories.FeatureFlagRepository, dbPath string, startedAt time.Time) *AdminService {
	return &AdminService{settings: settings, flags: flags, dbPath: dbPath, startedAt: startedAt}
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
	return &AdminDashboard{
		Settings:     settings,
		Flags:        flags,
		SystemHealth: s.SystemHealth(),
	}, nil
}

func (s *AdminService) UpdateApplication(ctx context.Context, appName, apiURL string) error {
	return s.settings.UpdateApplication(ctx, appName, apiURL)
}

func (s *AdminService) UpdateVersion(ctx context.Context, androidLatest, androidMin string, androidForce bool, iosLatest, iosMin string, iosForce bool) error {
	return s.settings.UpdateVersion(ctx, androidLatest, androidMin, androidForce, iosLatest, iosMin, iosForce)
}

func (s *AdminService) UpdateMaintenance(ctx context.Context, enabled bool, message string) error {
	return s.settings.UpdateMaintenance(ctx, enabled, message)
}

func (s *AdminService) UpdateBanner(ctx context.Context, enabled bool, message string) error {
	return s.settings.UpdateBanner(ctx, enabled, message)
}

func (s *AdminService) CreateFlag(ctx context.Context, key string, enabled bool) error {
	_, err := s.flags.Create(ctx, key, enabled)
	return err
}

func (s *AdminService) UpdateFlag(ctx context.Context, id int64, key string, enabled bool) error {
	return s.flags.Update(ctx, id, key, enabled)
}

func (s *AdminService) DeleteFlag(ctx context.Context, id int64) error {
	return s.flags.Delete(ctx, id)
}

func (s *AdminService) SystemHealth() SystemHealth {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	dbSize := s.sqliteFileSize()
	score := 100
	notes := make([]string, 0, 4)

	if mem.Alloc > 128*1024*1024 {
		score -= 20
		notes = append(notes, "Memory allocation is above 128 MB")
	}
	if mem.Sys > 256*1024*1024 {
		score -= 15
		notes = append(notes, "System memory usage is elevated")
	}
	if dbSize > 100*1024*1024 {
		score -= 15
		notes = append(notes, "SQLite database file is growing large")
	}
	if runtime.NumGoroutine() > 50 {
		score -= 10
		notes = append(notes, "Goroutine count is higher than expected")
	}

	if score < 0 {
		score = 0
	}

	status := "healthy"
	switch {
	case score >= 90:
		status = "excellent"
	case score >= 75:
		status = "good"
	case score >= 60:
		status = "fair"
	default:
		status = "needs attention"
	}

	return SystemHealth{
		Score:          score,
		Status:         status,
		Uptime:         time.Since(s.startedAt).Truncate(time.Second).String(),
		GoVersion:      runtime.Version(),
		NumCPU:         runtime.NumCPU(),
		NumGoroutine:   runtime.NumGoroutine(),
		AllocMB:        bytesToMB(mem.Alloc),
		TotalAllocMB:   bytesToMB(mem.TotalAlloc),
		SysMB:          bytesToMB(mem.Sys),
		SQLiteFileSize: humanFileSize(dbSize),
		SQLiteFilePath: s.dbPath,
		HealthNotes:    notes,
	}
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

func bytesToMB(v uint64) float64 {
	return float64(v) / (1024 * 1024)
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
