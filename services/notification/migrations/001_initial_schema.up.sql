-- Notification Service Initial Schema
-- Consolidated migration for pre-release

-- ============================================================================
-- Helper Functions
-- ============================================================================

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- Notification Templates Table
-- ============================================================================

CREATE TABLE IF NOT EXISTS notification_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    channel VARCHAR(20) NOT NULL,
    subject_template VARCHAR(500),
    body_template TEXT NOT NULL,
    version INT NOT NULL DEFAULT 1,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT templates_channel_check CHECK (channel IN ('sms', 'email', 'push', 'in_app')),
    CONSTRAINT templates_version_check CHECK (version > 0)
);

CREATE INDEX idx_templates_name ON notification_templates(name);
CREATE INDEX idx_templates_channel ON notification_templates(channel);

CREATE TRIGGER update_notification_templates_updated_at
    BEFORE UPDATE ON notification_templates
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE notification_templates IS 'Reusable notification templates with variable placeholders';
COMMENT ON COLUMN notification_templates.name IS 'Unique template identifier (e.g., otp_sms, transaction_alert_email)';
COMMENT ON COLUMN notification_templates.version IS 'Template version for tracking changes';

-- ============================================================================
-- Notifications Table
-- ============================================================================

CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID,
    channel VARCHAR(20) NOT NULL,
    type VARCHAR(50) NOT NULL,
    priority VARCHAR(20) NOT NULL DEFAULT 'normal',
    recipient VARCHAR(255) NOT NULL,
    subject VARCHAR(500),
    body TEXT NOT NULL,
    template_id UUID REFERENCES notification_templates(id) ON DELETE SET NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'queued',
    correlation_id VARCHAR(100),
    source_service VARCHAR(50) NOT NULL,
    metadata JSONB,
    retry_count INT NOT NULL DEFAULT 0,
    failure_reason TEXT,
    queued_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    sent_at TIMESTAMP WITH TIME ZONE,
    delivered_at TIMESTAMP WITH TIME ZONE,
    failed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT notifications_channel_check CHECK (channel IN ('sms', 'email', 'push', 'in_app')),
    CONSTRAINT notifications_priority_check CHECK (priority IN ('critical', 'high', 'normal', 'low')),
    CONSTRAINT notifications_status_check CHECK (status IN ('queued', 'sent', 'delivered', 'failed')),
    CONSTRAINT notifications_retry_count_check CHECK (retry_count >= 0 AND retry_count <= 10)
);

CREATE INDEX idx_notifications_user_id ON notifications(user_id) WHERE user_id IS NOT NULL;
CREATE INDEX idx_notifications_channel ON notifications(channel);
CREATE INDEX idx_notifications_type ON notifications(type);
CREATE INDEX idx_notifications_status ON notifications(status);
CREATE INDEX idx_notifications_correlation_id ON notifications(correlation_id) WHERE correlation_id IS NOT NULL;
CREATE INDEX idx_notifications_source_service ON notifications(source_service);
CREATE INDEX idx_notifications_created_at ON notifications(created_at DESC);
CREATE INDEX idx_notifications_priority_status ON notifications(priority, status) WHERE status = 'queued';

CREATE UNIQUE INDEX idx_notifications_correlation_id_unique ON notifications(correlation_id) WHERE correlation_id IS NOT NULL;

CREATE TRIGGER update_notifications_updated_at
    BEFORE UPDATE ON notifications
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE notifications IS 'Stores all outbound notifications for simulation and audit';
COMMENT ON COLUMN notifications.correlation_id IS 'For idempotency - prevents duplicate notifications';
COMMENT ON COLUMN notifications.priority IS 'Critical (OTP) processed first, then high, normal, low';
COMMENT ON COLUMN notifications.retry_count IS 'Number of retry attempts (max 10)';

-- ============================================================================
-- Seed Data: Notification Templates
-- ============================================================================

-- OTP SMS Template
INSERT INTO notification_templates (name, channel, subject_template, body_template, version)
VALUES (
    'otp_sms',
    'sms',
    '',
    'Your Nivo Money OTP is {{otp}}. Valid for {{validity_minutes}} minutes. Do not share this code. - Nivo Money',
    1
);

-- Transaction Alert SMS Template
INSERT INTO notification_templates (name, channel, subject_template, body_template, version)
VALUES (
    'transaction_alert_sms',
    'sms',
    '',
    'Txn Alert: Rs {{amount}} {{transaction_type}} on {{date}}. Wallet bal: Rs {{balance}}. Ref: {{transaction_id}}. - Nivo Money',
    1
);

-- Welcome Email Template
INSERT INTO notification_templates (name, channel, subject_template, body_template, version)
VALUES (
    'welcome_email',
    'email',
    'Welcome to Nivo Money, {{user_name}}!',
    'Dear {{user_name}},

Welcome to Nivo Money! We''re excited to have you on board.

Your account has been successfully created. You can now:
- Create wallets
- Send and receive money
- Track your transactions

Get started by creating your first wallet.

If you have any questions, our support team is here to help.

Best regards,
The Nivo Money Team',
    1
);

-- Transaction Alert Email Template
INSERT INTO notification_templates (name, channel, subject_template, body_template, version)
VALUES (
    'transaction_alert_email',
    'email',
    'Transaction Alert: {{transaction_type}} of ₹{{amount}}',
    'Dear {{user_name}},

This is to inform you about a recent transaction on your Nivo Money account:

Transaction Details:
- Type: {{transaction_type}}
- Amount: ₹{{amount}}
- Date: {{date}}
- Reference: {{transaction_id}}
- Description: {{description}}

Current Wallet Balance: ₹{{balance}}

If you did not authorize this transaction, please contact our support team immediately.

Best regards,
The Nivo Money Team',
    1
);

-- KYC Update Email Template
INSERT INTO notification_templates (name, channel, subject_template, body_template, version)
VALUES (
    'kyc_update_email',
    'email',
    'KYC Status Update - {{kyc_status}}',
    'Dear {{user_name}},

Your KYC (Know Your Customer) verification status has been updated:

Status: {{kyc_status}}

{{#if approved}}
Congratulations! Your account is now fully verified. You can now enjoy all features of Nivo Money without any limits.
{{else}}
{{#if rejected}}
Unfortunately, we were unable to verify your documents. Reason: {{rejection_reason}}

Please submit valid documents to complete your verification.
{{else}}
Your KYC verification is currently under review. We''ll notify you once the review is complete.
{{/if}}
{{/if}}

Thank you for choosing Nivo Money.

Best regards,
The Nivo Money Team',
    1
);

-- Account Alert Email Template
INSERT INTO notification_templates (name, channel, subject_template, body_template, version)
VALUES (
    'account_alert_email',
    'email',
    'Important Account Alert',
    'Dear {{user_name}},

This is an important alert regarding your Nivo Money account:

Alert Type: {{alert_type}}
Message: {{message}}
Date: {{date}}

If this activity was not authorized by you, please contact our support team immediately.

Best regards,
The Nivo Money Team',
    1
);

-- Wallet Created Push Notification Template
INSERT INTO notification_templates (name, channel, subject_template, body_template, version)
VALUES (
    'wallet_created_push',
    'push',
    'Wallet Created Successfully',
    'Your new {{currency}} wallet has been created. Start sending and receiving money now!',
    1
);

-- Transaction Alert Push Notification Template
INSERT INTO notification_templates (name, channel, subject_template, body_template, version)
VALUES (
    'transaction_alert_push',
    'push',
    'Transaction: ₹{{amount}}',
    '{{transaction_type}} of ₹{{amount}} completed. Balance: ₹{{balance}}',
    1
);

-- Security Alert Push Notification Template
INSERT INTO notification_templates (name, channel, subject_template, body_template, version)
VALUES (
    'security_alert_push',
    'push',
    'Security Alert',
    '{{message}}. If this wasn''t you, please secure your account immediately.',
    1
);

-- In-App Welcome Message Template
INSERT INTO notification_templates (name, channel, subject_template, body_template, version)
VALUES (
    'welcome_inapp',
    'in_app',
    'Welcome to Nivo Money!',
    'Hi {{user_name}}! Welcome to Nivo Money. Create your first wallet to get started.',
    1
);

-- Welcome SMS Template
INSERT INTO notification_templates (name, channel, subject_template, body_template, version)
VALUES (
    'welcome_sms',
    'sms',
    '',
    'Welcome to Nivo Money, {{full_name}}! Your account is ready. Create your first wallet to get started. - Nivo Money',
    1
);

-- KYC Approved Email Template
INSERT INTO notification_templates (name, channel, subject_template, body_template, version)
VALUES (
    'kyc_approved_email',
    'email',
    'KYC Verification Approved',
    'Dear {{full_name}},

Congratulations! Your KYC verification has been approved.

Your account is now fully verified and you can enjoy all features of Nivo Money without any limits:
- Create multiple wallets
- Send and receive money
- Higher transaction limits
- Access to all premium features

Thank you for choosing Nivo Money.

Best regards,
The Nivo Money Team',
    1
);

-- KYC Approved SMS Template
INSERT INTO notification_templates (name, channel, subject_template, body_template, version)
VALUES (
    'kyc_approved_sms',
    'sms',
    '',
    'Good news {{full_name}}! Your KYC verification is approved. Your account is now fully activated. - Nivo Money',
    1
);

-- KYC Rejected Email Template
INSERT INTO notification_templates (name, channel, subject_template, body_template, version)
VALUES (
    'kyc_rejected_email',
    'email',
    'KYC Verification - Action Required',
    'Dear {{full_name}},

We were unable to verify your KYC documents.

Reason: {{reason}}

Please log in to your account and resubmit valid documents to complete your verification. If you have any questions, please contact our support team.

Thank you for your patience.

Best regards,
The Nivo Money Team',
    1
);

-- KYC Rejected SMS Template
INSERT INTO notification_templates (name, channel, subject_template, body_template, version)
VALUES (
    'kyc_rejected_sms',
    'sms',
    '',
    'KYC verification rejected. Reason: {{reason}}. Please resubmit documents. - Nivo Money',
    1
);

-- Wallet Created Email Template
INSERT INTO notification_templates (name, channel, subject_template, body_template, version)
VALUES (
    'wallet_created_email',
    'email',
    'New Wallet Created',
    'Dear Customer,

Your new {{wallet_type}} wallet has been created successfully.

Wallet Details:
- Type: {{wallet_type}}
- Currency: {{currency}}
- Wallet ID: {{wallet_id}}
- Status: Inactive (pending KYC verification)

Once your KYC is verified, your wallet will be automatically activated and ready to use.

Best regards,
The Nivo Money Team',
    1
);

-- Wallet Activated Email Template
INSERT INTO notification_templates (name, channel, subject_template, body_template, version)
VALUES (
    'wallet_activated_email',
    'email',
    'Wallet Activated',
    'Dear Customer,

Your {{wallet_type}} wallet has been activated and is ready to use!

Wallet Details:
- Type: {{wallet_type}}
- Currency: {{currency}}
- Wallet ID: {{wallet_id}}

You can now:
- Receive money
- Send money to other users
- Make deposits and withdrawals

Best regards,
The Nivo Money Team',
    1
);

-- Wallet Activated SMS Template
INSERT INTO notification_templates (name, channel, subject_template, body_template, version)
VALUES (
    'wallet_activated_sms',
    'sms',
    '',
    'Your {{wallet_type}} wallet ({{currency}}) is now active! Start sending and receiving money. - Nivo Money',
    1
);

-- Transaction Created Email Template
INSERT INTO notification_templates (name, channel, subject_template, body_template, version)
VALUES (
    'transaction_created_email',
    'email',
    'Transaction Initiated: {{transaction_type}}',
    'Dear Customer,

A new transaction has been initiated on your Nivo Money account:

Transaction Details:
- Type: {{transaction_type}}
- Amount: {{currency}} {{amount}}
- Transaction ID: {{transaction_id}}
- Description: {{description}}
- Status: Pending

You will receive another notification once the transaction is completed.

If you did not authorize this transaction, please contact our support team immediately.

Best regards,
The Nivo Money Team',
    1
);

-- Transaction Created SMS Template
INSERT INTO notification_templates (name, channel, subject_template, body_template, version)
VALUES (
    'transaction_created_sms',
    'sms',
    '',
    'Transaction initiated: {{transaction_type}} of {{currency}} {{amount}}. Ref: {{transaction_id}}. - Nivo Money',
    1
);
