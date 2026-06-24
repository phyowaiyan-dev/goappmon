package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/phyowaiyan-dev/goappmon/internal/models"
)

type AuditLogRepository struct {
	db DBTX
}

type AuditLogFilter struct {
	Query  string
	Limit  int
	Offset int
}

type AuditLogPage struct {
	Logs  []models.AuditLog
	Total int64
}

func NewAuditLogRepository(db DBTX) *AuditLogRepository {
	return &AuditLogRepository{db: db}
}

func (r *AuditLogRepository) Create(ctx context.Context, log models.AuditLog) (int64, error) {
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO audit_logs (
			actor_admin_id, action, entity_type, entity_id, before_json, after_json, ip, user_agent, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		log.ActorAdminID,
		log.Action,
		log.EntityType,
		log.EntityID,
		log.BeforeJSON,
		log.AfterJSON,
		log.IP,
		log.UserAgent,
		log.CreatedAt.Unix(),
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *AuditLogRepository) ListRecent(ctx context.Context, limit int) ([]models.AuditLog, error) {
	page, err := r.Search(ctx, AuditLogFilter{Limit: limit})
	if err != nil {
		return nil, err
	}
	return page.Logs, nil
}

func (r *AuditLogRepository) Search(ctx context.Context, filter AuditLogFilter) (AuditLogPage, error) {
	if filter.Limit <= 0 {
		filter.Limit = 10
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	where, args := buildAuditWhere(filter)
	query := fmt.Sprintf(`
		SELECT al.id, al.actor_admin_id, COALESCE(a.name, ''), al.action, al.entity_type, al.entity_id,
		       al.before_json, al.after_json, al.ip, al.user_agent, al.created_at
		FROM audit_logs al
		LEFT JOIN admins a ON a.id = al.actor_admin_id
		%s
		ORDER BY al.created_at DESC, al.id DESC
		LIMIT ? OFFSET ?
	`, where)
	argsWithLimit := append(args, filter.Limit, filter.Offset)
	rows, err := r.db.QueryContext(ctx, query, argsWithLimit...)
	if err != nil {
		return AuditLogPage{}, err
	}
	defer rows.Close()

	logs := make([]models.AuditLog, 0)
	for rows.Next() {
		entry, err := scanAuditLog(rows)
		if err != nil {
			return AuditLogPage{}, err
		}
		logs = append(logs, *entry)
	}
	if err := rows.Err(); err != nil {
		return AuditLogPage{}, err
	}

	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM audit_logs al
		LEFT JOIN admins a ON a.id = al.actor_admin_id
		%s
	`, where)
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return AuditLogPage{}, err
	}

	return AuditLogPage{Logs: logs, Total: total}, nil
}

func buildAuditWhere(filter AuditLogFilter) (string, []any) {
	clauses := make([]string, 0, 2)
	args := make([]any, 0, 2)
	clauses = append(clauses, "1=1")

	if q := strings.TrimSpace(filter.Query); q != "" {
		like := "%" + strings.ToLower(q) + "%"
		clauses = append(clauses, "(LOWER(al.action) LIKE ? OR LOWER(al.entity_type) LIKE ? OR LOWER(al.entity_id) LIKE ? OR LOWER(al.ip) LIKE ? OR LOWER(al.user_agent) LIKE ? OR LOWER(COALESCE(a.name, '')) LIKE ?)")
		for i := 0; i < 6; i++ {
			args = append(args, like)
		}
	}
	return "WHERE " + strings.Join(clauses, " AND "), args
}

func scanAuditLog(rows *sql.Rows) (*models.AuditLog, error) {
	var entry models.AuditLog
	var createdAt int64
	if err := rows.Scan(
		&entry.ID,
		&entry.ActorAdminID,
		&entry.ActorName,
		&entry.Action,
		&entry.EntityType,
		&entry.EntityID,
		&entry.BeforeJSON,
		&entry.AfterJSON,
		&entry.IP,
		&entry.UserAgent,
		&createdAt,
	); err != nil {
		return nil, err
	}
	entry.CreatedAt = time.Unix(createdAt, 0).UTC()
	return &entry, nil
}
