-- Remove user workflow templates

DELETE FROM notification_templates WHERE name IN (
    'welcome_sms',
    'kyc_approved_email',
    'kyc_approved_sms',
    'kyc_rejected_email',
    'kyc_rejected_sms',
    'wallet_created_email',
    'wallet_activated_email',
    'wallet_activated_sms',
    'transaction_created_email',
    'transaction_created_sms'
);
