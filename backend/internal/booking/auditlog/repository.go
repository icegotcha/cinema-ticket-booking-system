package booking_auditlog

import (
	"github.com/icegotcha/cinema-ticket-booking-system/backend/internal/database"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type BookingAuditLogRepository struct {
	*database.Repository[BookingAuditLog]
}

func NewBookingAuditLogRepository(db *mongo.Database) *BookingAuditLogRepository {
	return &BookingAuditLogRepository{
		Repository: database.NewRepository[BookingAuditLog](db, "booking_audit_logs"),
	}
}
