# Security Policy

## Scope

Nivo is a **portfolio demonstration project** — not a production bank. It handles no real money or real customer data. That said, the codebase aims to demonstrate production-grade security practices.

## Reporting a Vulnerability

If you find a security issue, please report it responsibly:

1. **Do not** open a public GitHub issue
2. Email: **security@nivomoney.com** (or open a [private security advisory](https://github.com/vnykmshr/nivomoney/security/advisories/new))
3. Include: description, steps to reproduce, and potential impact

You should receive a response within 72 hours.

## Security Measures

This project implements:

- JWT authentication with bcrypt password hashing
- Role-based access control (RBAC) with granular permissions
- CSRF protection on state-changing requests
- Rate limiting on authentication and API endpoints
- Content Security Policy headers
- HTTPS-only with TLS 1.2/1.3
- Docker containers with dropped capabilities and no-new-privileges
- Input validation at service boundaries
- No hardcoded secrets — SOPS/age encryption for production config

For detailed frontend security architecture, see [`frontend/SECURITY.md`](../frontend/SECURITY.md).

## Supported Versions

| Version | Supported |
|---------|-----------|
| main    | Current   |
