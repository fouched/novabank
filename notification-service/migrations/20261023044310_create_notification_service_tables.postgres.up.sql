-- 1. Create Custom PostgreSQL Enum Types
CREATE TYPE notification_type_enum   AS ENUM ('EMAIL', 'SMS');
CREATE TYPE notification_status_enum AS ENUM ('SENT', 'FAILED');

-- 2. Create Notifications Storage Table
CREATE TABLE notifications (
    id                    BIGSERIAL PRIMARY KEY,
    type                  notification_type_enum NOT NULL,
    status                notification_status_enum NOT NULL,
    message               TEXT NOT NULL,
    recipient_email       TEXT,
    recipient_phone       TEXT,
    subject               TEXT,
    transaction_reference TEXT,
    created_at            TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 3. Add High-Performance Operational Indices
CREATE INDEX idx_notifications_recipient_email ON notifications(recipient_email);
CREATE INDEX idx_notifications_tx_ref ON notifications(transaction_reference);
