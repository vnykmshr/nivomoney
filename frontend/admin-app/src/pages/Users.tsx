/**
 * Users Page - Admin User Management
 * Search and manage users
 */

import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { AdminLayout } from '../components';
import { adminApi } from '../lib/adminApi';
import type { User } from '@nivo/shared';
import {
  Card,
  CardTitle,
  Button,
  Input,
  Alert,
  Badge,
  Skeleton,
  FormField,
} from '@nivo/shared';
import { getStatusVariant, getKYCStatusVariant } from '@nivo/shared';

export function Users() {
  const navigate = useNavigate();
  const [searchQuery, setSearchQuery] = useState('');
  const [searchResults, setSearchResults] = useState<User[]>([]);
  const [isSearching, setIsSearching] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [hasSearched, setHasSearched] = useState(false);

  const handleSearch = async () => {
    if (!searchQuery.trim()) {
      setError('Please enter a search term');
      return;
    }

    if (searchQuery.trim().length < 2) {
      setError('Please enter at least 2 characters to search');
      return;
    }

    setIsSearching(true);
    setError(null);
    setHasSearched(true);

    try {
      const results = await adminApi.searchUsers(searchQuery);
      setSearchResults(results);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Search failed');
      setSearchResults([]);
    } finally {
      setIsSearching(false);
    }
  };

  return (
    <AdminLayout title="User Management">
      <div className="space-y-6">
        {/* Error Alert */}
        {error && (
          <Alert variant="error" onDismiss={() => setError(null)}>
            {error}
          </Alert>
        )}

        {/* Search Card */}
        <Card>
          <CardTitle className="mb-4">Search Users</CardTitle>
          <div className="flex gap-3">
            <FormField
              label="Search by email, phone, or name"
              htmlFor="user-search"
              className="flex-1"
            >
              <Input
                id="user-search"
                value={searchQuery}
                onChange={e => setSearchQuery(e.target.value)}
                onKeyDown={e => e.key === 'Enter' && handleSearch()}
                placeholder="Enter email, phone, or name..."
              />
            </FormField>
          </div>
          <Button
            onClick={handleSearch}
            loading={isSearching}
            className="mt-4 w-full sm:w-auto"
          >
            Search
          </Button>
        </Card>

        {/* Loading State */}
        {isSearching && (
          <div className="space-y-4">
            {[1, 2, 3].map(i => (
              <Card key={i}>
                <div className="flex justify-between items-start">
                  <div className="flex-1">
                    <Skeleton className="h-6 w-48 mb-2" />
                    <Skeleton className="h-4 w-56 mb-1" />
                    <Skeleton className="h-4 w-40" />
                  </div>
                  <Skeleton className="h-6 w-20" />
                </div>
              </Card>
            ))}
          </div>
        )}

        {/* Results */}
        {!isSearching && hasSearched && searchResults.length > 0 && (
          <div className="space-y-4">
            <h3 className="text-lg font-semibold text-[var(--text-primary)]">
              Found {searchResults.length} user{searchResults.length !== 1 ? 's' : ''}
            </h3>

            {searchResults.map(user => (
              <Card
                key={user.id}
                className="cursor-pointer hover:shadow-md transition-shadow"
                onClick={() => navigate(`/users/${user.id}`)}
              >
                <div className="flex justify-between items-start">
                  <div className="flex-1">
                    <h3 className="text-lg font-semibold text-[var(--text-primary)]">
                      {user.full_name}
                    </h3>
                    <p className="text-sm text-[var(--text-secondary)]">{user.email}</p>
                    <p className="text-sm text-[var(--text-secondary)]">{user.phone}</p>
                    <p className="text-xs text-[var(--text-muted)] mt-1 font-mono">
                      ID: {user.id.slice(0, 8)}...
                    </p>
                  </div>
                  <div className="flex flex-col items-end gap-2">
                    <Badge variant={getStatusVariant(user.status)}>
                      {user.status.toUpperCase()}
                    </Badge>
                    {user.kyc && (
                      <Badge variant={getKYCStatusVariant(user.kyc.status)}>
                        KYC: {user.kyc.status}
                      </Badge>
                    )}
                    <Button size="sm" variant="secondary">View Details</Button>
                  </div>
                </div>
              </Card>
            ))}
          </div>
        )}

        {/* Empty State - No Results */}
        {!isSearching && hasSearched && searchResults.length === 0 && (
          <Card className="text-center py-12">
            <div className="w-16 h-16 mx-auto mb-4 rounded-full bg-[var(--surface-secondary)] flex items-center justify-center">
              <svg className="w-8 h-8 text-[var(--text-muted)]" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
              </svg>
            </div>
            <CardTitle className="mb-2">No Users Found</CardTitle>
            <p className="text-[var(--text-muted)]">
              No users match your search criteria. Try a different search term.
            </p>
          </Card>
        )}

        {/* Initial State - No Search Yet */}
        {!isSearching && !hasSearched && (
          <Card className="text-center py-12">
            <div className="w-16 h-16 mx-auto mb-4 rounded-full bg-[var(--surface-brand-subtle)] flex items-center justify-center">
              <svg className="w-8 h-8 text-[var(--interactive-primary)]" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
              </svg>
            </div>
            <CardTitle className="mb-2">Search for Users</CardTitle>
            <p className="text-[var(--text-muted)]">
              Enter an email, phone number, or name to find users
            </p>
          </Card>
        )}
      </div>
    </AdminLayout>
  );
}
