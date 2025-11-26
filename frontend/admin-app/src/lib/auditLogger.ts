/**
 * Admin Audit Logger
 * Tracks all admin actions for compliance and security
 */

import { AuditLogger, type Environment } from '@nivo/shared';

// Determine environment
const env: Environment = (import.meta.env.MODE === 'production') ? 'production' : 'development';

// Create singleton audit logger instance
export const auditLogger = new AuditLogger(true, env);

// Convenience functions for common admin actions
export const logAdminAction = {
  login: (userId: string, userName: string) => {
    auditLogger.log({
      userId,
      userName,
      action: 'LOGIN',
      resource: 'auth',
      details: {
        timestamp: new Date().toISOString(),
      },
    });
  },

  logout: (userId: string, userName: string) => {
    auditLogger.log({
      userId,
      userName,
      action: 'LOGOUT',
      resource: 'auth',
      details: {
        timestamp: new Date().toISOString(),
      },
    });
  },

  verifyKYC: (adminUserId: string, adminUserName: string, targetUserId: string, targetUserName: string) => {
    auditLogger.log({
      userId: adminUserId,
      userName: adminUserName,
      action: 'VERIFY_KYC',
      resource: `user:${targetUserId}`,
      details: {
        targetUserId,
        targetUserName,
        decision: 'approved',
      },
    });
  },

  rejectKYC: (adminUserId: string, adminUserName: string, targetUserId: string, targetUserName: string, reason: string) => {
    auditLogger.log({
      userId: adminUserId,
      userName: adminUserName,
      action: 'REJECT_KYC',
      resource: `user:${targetUserId}`,
      details: {
        targetUserId,
        targetUserName,
        decision: 'rejected',
        reason,
      },
    });
  },

  viewPendingKYCs: (adminUserId: string, adminUserName: string, count: number) => {
    auditLogger.log({
      userId: adminUserId,
      userName: adminUserName,
      action: 'VIEW_PENDING_KYCS',
      resource: 'kyc_list',
      details: {
        count,
      },
    });
  },

  viewDashboard: (adminUserId: string, adminUserName: string) => {
    auditLogger.log({
      userId: adminUserId,
      userName: adminUserName,
      action: 'VIEW_DASHBOARD',
      resource: 'admin_dashboard',
      details: {},
    });
  },

  searchUser: (adminUserId: string, adminUserName: string, searchQuery: string) => {
    auditLogger.log({
      userId: adminUserId,
      userName: adminUserName,
      action: 'SEARCH_USER',
      resource: 'user_search',
      details: {
        searchQuery,
      },
    });
  },

  viewUserDetails: (adminUserId: string, adminUserName: string, targetUserId: string) => {
    auditLogger.log({
      userId: adminUserId,
      userName: adminUserName,
      action: 'VIEW_USER_DETAILS',
      resource: `user:${targetUserId}`,
      details: {
        targetUserId,
      },
    });
  },

  suspendUser: (adminUserId: string, adminUserName: string, targetUserId: string, reason: string) => {
    auditLogger.log({
      userId: adminUserId,
      userName: adminUserName,
      action: 'SUSPEND_USER',
      resource: `user:${targetUserId}`,
      details: {
        targetUserId,
        reason,
      },
    });
  },

  viewTransaction: (adminUserId: string, adminUserName: string, transactionId: string) => {
    auditLogger.log({
      userId: adminUserId,
      userName: adminUserName,
      action: 'VIEW_TRANSACTION',
      resource: `transaction:${transactionId}`,
      details: {
        transactionId,
      },
    });
  },
};
