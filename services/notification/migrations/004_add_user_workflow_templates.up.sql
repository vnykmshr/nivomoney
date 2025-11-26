-- Add notification templates for user workflow integration

-- Welcome SMS Template
INSERT INTO notification_templates (id, name, channel, subject_template, body_template, version, created_at, updated_at)
VALUES (
    gen_random_uuid(),
    'welcome_sms',
    'sms',
    '',
    'Welcome to Nivo Money, {{full_name}}! Your account is ready. Create your first wallet to get started. - Nivo Money',
    1,
    NOW(),
    NOW()
);

-- KYC Approved Email Template
INSERT INTO notification_templates (id, name, channel, subject_template, body_template, version, created_at, updated_at)
VALUES (
    gen_random_uuid(),
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
    1,
    NOW(),
    NOW()
);

-- KYC Approved SMS Template
INSERT INTO notification_templates (id, name, channel, subject_template, body_template, version, created_at, updated_at)
VALUES (
    gen_random_uuid(),
    'kyc_approved_sms',
    'sms',
    '',
    'Good news {{full_name}}! Your KYC verification is approved. Your account is now fully activated. - Nivo Money',
    1,
    NOW(),
    NOW()
);

-- KYC Rejected Email Template
INSERT INTO notification_templates (id, name, channel, subject_template, body_template, version, created_at, updated_at)
VALUES (
    gen_random_uuid(),
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
    1,
    NOW(),
    NOW()
);

-- KYC Rejected SMS Template
INSERT INTO notification_templates (id, name, channel, subject_template, body_template, version, created_at, updated_at)
VALUES (
    gen_random_uuid(),
    'kyc_rejected_sms',
    'sms',
    '',
    'KYC verification rejected. Reason: {{reason}}. Please resubmit documents. - Nivo Money',
    1,
    NOW(),
    NOW()
);

-- Wallet Created Email Template
INSERT INTO notification_templates (id, name, channel, subject_template, body_template, version, created_at, updated_at)
VALUES (
    gen_random_uuid(),
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
    1,
    NOW(),
    NOW()
);

-- Wallet Activated Email Template
INSERT INTO notification_templates (id, name, channel, subject_template, body_template, version, created_at, updated_at)
VALUES (
    gen_random_uuid(),
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
    1,
    NOW(),
    NOW()
);

-- Wallet Activated SMS Template
INSERT INTO notification_templates (id, name, channel, subject_template, body_template, version, created_at, updated_at)
VALUES (
    gen_random_uuid(),
    'wallet_activated_sms',
    'sms',
    '',
    'Your {{wallet_type}} wallet ({{currency}}) is now active! Start sending and receiving money. - Nivo Money',
    1,
    NOW(),
    NOW()
);

-- Transaction Created Email Template
INSERT INTO notification_templates (id, name, channel, subject_template, body_template, version, created_at, updated_at)
VALUES (
    gen_random_uuid(),
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
    1,
    NOW(),
    NOW()
);

-- Transaction Created SMS Template
INSERT INTO notification_templates (id, name, channel, subject_template, body_template, version, created_at, updated_at)
VALUES (
    gen_random_uuid(),
    'transaction_created_sms',
    'sms',
    '',
    'Transaction initiated: {{transaction_type}} of {{currency}} {{amount}}. Ref: {{transaction_id}}. - Nivo Money',
    1,
    NOW(),
    NOW()
);
