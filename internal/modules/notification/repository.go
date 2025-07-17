package notification

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type DBTX interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}
type notificationRepository struct {
}

func NewNotificationRepository(conn DBTX) NotificationRepository {
	return &notificationRepository{}
}

// Placeholder implementations for future background task processing
func (r *notificationRepository) InsertNotificationTask(ctx context.Context, task *NotificationTask) error {
	// TODO: Implement when we add notification tasks table
	return nil
}

func (r *notificationRepository) GetPendingTasks(ctx context.Context, limit int) ([]*NotificationTask, error) {
	// TODO: Implement when we add notification tasks table
	return nil, nil
}

func (r *notificationRepository) DeleteTask(ctx context.Context, id string) error {
	// TODO: Implement when we add notification tasks table
	return nil
}
