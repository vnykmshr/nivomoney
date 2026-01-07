/**
 * Pagination Component
 * Navigate through paginated data with page numbers and navigation buttons
 */

import { useMemo, useState, useCallback } from 'react';
import type { HTMLAttributes } from 'react';
import { cn } from '../../lib/utils';

export interface PaginationProps extends HTMLAttributes<HTMLElement> {
  /** Current page (1-indexed) */
  currentPage: number;
  /** Total number of pages */
  totalPages: number;
  /** Callback when page changes */
  onPageChange: (page: number) => void;
  /** Number of page buttons to show on each side of current page */
  siblingCount?: number;
  /** Show first/last page buttons */
  showEndButtons?: boolean;
  /** Custom labels */
  labels?: {
    previous?: string;
    next?: string;
    first?: string;
    last?: string;
  };
  /** Size variant */
  size?: 'sm' | 'md' | 'lg';
  /** Disable all interactions */
  disabled?: boolean;
}

const DOTS = '...';

function usePaginationRange(
  currentPage: number,
  totalPages: number,
  siblingCount: number
): (number | string)[] {
  return useMemo(() => {
    // Total slots = siblings on each side + first + last + current + 2 dots
    const totalPageNumbers = siblingCount * 2 + 5;

    // If we have fewer pages than slots, show all pages
    if (totalPages <= totalPageNumbers) {
      return Array.from({ length: totalPages }, (_, i) => i + 1);
    }

    const leftSiblingIndex = Math.max(currentPage - siblingCount, 1);
    const rightSiblingIndex = Math.min(currentPage + siblingCount, totalPages);

    const showLeftDots = leftSiblingIndex > 2;
    const showRightDots = rightSiblingIndex < totalPages - 1;

    if (!showLeftDots && showRightDots) {
      // Show more pages on the left
      const leftRange = Array.from({ length: 3 + siblingCount * 2 }, (_, i) => i + 1);
      return [...leftRange, DOTS, totalPages];
    }

    if (showLeftDots && !showRightDots) {
      // Show more pages on the right
      const rightRange = Array.from(
        { length: 3 + siblingCount * 2 },
        (_, i) => totalPages - (3 + siblingCount * 2) + i + 1
      );
      return [1, DOTS, ...rightRange];
    }

    // Show dots on both sides
    const middleRange = Array.from(
      { length: rightSiblingIndex - leftSiblingIndex + 1 },
      (_, i) => leftSiblingIndex + i
    );
    return [1, DOTS, ...middleRange, DOTS, totalPages];
  }, [currentPage, totalPages, siblingCount]);
}

export function Pagination({
  currentPage,
  totalPages,
  onPageChange,
  siblingCount = 1,
  showEndButtons = false,
  labels = {},
  size = 'md',
  disabled = false,
  className,
  ...props
}: PaginationProps) {
  const paginationRange = usePaginationRange(currentPage, totalPages, siblingCount);

  const {
    previous = 'Previous',
    next = 'Next',
    first = 'First',
    last = 'Last',
  } = labels;

  const sizeClasses = {
    sm: 'text-xs px-2 py-1 min-w-[28px]',
    md: 'text-sm px-3 py-2 min-w-[36px]',
    lg: 'text-base px-4 py-2.5 min-w-[44px]',
  }[size];

  const isFirstPage = currentPage === 1;
  const isLastPage = currentPage === totalPages;

  if (totalPages <= 1) return null;

  const baseButtonClass = cn(
    'inline-flex items-center justify-center font-medium',
    'rounded-[var(--radius-button)]',
    'transition-colors focus:outline-none focus:[box-shadow:var(--focus-ring)]',
    sizeClasses
  );

  const navButtonClass = cn(
    baseButtonClass,
    'text-[var(--text-secondary)] hover:text-[var(--text-primary)]',
    'hover:bg-[var(--interactive-secondary)]',
    'disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:bg-transparent'
  );

  const pageButtonClass = (isActive: boolean) =>
    cn(
      baseButtonClass,
      isActive
        ? 'bg-[var(--interactive-primary)] text-[var(--text-inverse)]'
        : 'text-[var(--text-secondary)] hover:bg-[var(--interactive-secondary)] hover:text-[var(--text-primary)]',
      disabled && 'pointer-events-none opacity-50'
    );

  return (
    <nav
      role="navigation"
      aria-label="Pagination"
      className={cn('flex items-center gap-1', className)}
      {...props}
    >
      {/* First page button */}
      {showEndButtons && (
        <button
          onClick={() => onPageChange(1)}
          disabled={disabled || isFirstPage}
          className={navButtonClass}
          aria-label={first}
        >
          <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 19l-7-7 7-7m8 14l-7-7 7-7" />
          </svg>
        </button>
      )}

      {/* Previous button */}
      <button
        onClick={() => onPageChange(currentPage - 1)}
        disabled={disabled || isFirstPage}
        className={navButtonClass}
        aria-label={previous}
      >
        <svg className="w-4 h-4 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
        </svg>
        <span className="hidden sm:inline">{previous}</span>
      </button>

      {/* Page numbers */}
      <div className="flex items-center gap-1">
        {paginationRange.map((page, index) => {
          if (page === DOTS) {
            return (
              <span
                key={`dots-${index}`}
                className={cn(
                  'px-2 text-[var(--text-muted)]',
                  disabled && 'opacity-50'
                )}
                aria-hidden="true"
              >
                {DOTS}
              </span>
            );
          }

          const pageNumber = page as number;
          const isActive = pageNumber === currentPage;

          return (
            <button
              key={pageNumber}
              onClick={() => onPageChange(pageNumber)}
              disabled={disabled}
              className={pageButtonClass(isActive)}
              aria-label={`Page ${pageNumber}`}
              aria-current={isActive ? 'page' : undefined}
            >
              {pageNumber}
            </button>
          );
        })}
      </div>

      {/* Next button */}
      <button
        onClick={() => onPageChange(currentPage + 1)}
        disabled={disabled || isLastPage}
        className={navButtonClass}
        aria-label={next}
      >
        <span className="hidden sm:inline">{next}</span>
        <svg className="w-4 h-4 ml-1" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
        </svg>
      </button>

      {/* Last page button */}
      {showEndButtons && (
        <button
          onClick={() => onPageChange(totalPages)}
          disabled={disabled || isLastPage}
          className={navButtonClass}
          aria-label={last}
        >
          <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 5l7 7-7 7M5 5l7 7-7 7" />
          </svg>
        </button>
      )}
    </nav>
  );
}

/* ============================================
   PAGINATION INFO COMPONENT
   ============================================ */

export interface PaginationInfoProps extends HTMLAttributes<HTMLDivElement> {
  /** Current page (1-indexed) */
  currentPage: number;
  /** Items per page */
  pageSize: number;
  /** Total number of items */
  totalItems: number;
  /** Custom format function */
  format?: (start: number, end: number, total: number) => string;
}

export function PaginationInfo({
  currentPage,
  pageSize,
  totalItems,
  format,
  className,
  ...props
}: PaginationInfoProps) {
  const start = (currentPage - 1) * pageSize + 1;
  const end = Math.min(currentPage * pageSize, totalItems);

  const defaultFormat = (s: number, e: number, t: number) =>
    `Showing ${s} to ${e} of ${t} results`;

  const text = (format ?? defaultFormat)(start, end, totalItems);

  return (
    <div
      className={cn('text-sm text-[var(--text-muted)]', className)}
      {...props}
    >
      {text}
    </div>
  );
}

/* ============================================
   PAGE SIZE SELECTOR
   ============================================ */

export interface PageSizeSelectorProps extends HTMLAttributes<HTMLDivElement> {
  /** Current page size */
  pageSize: number;
  /** Available page size options */
  options?: number[];
  /** Callback when page size changes */
  onPageSizeChange: (size: number) => void;
  /** Label text */
  label?: string;
}

export function PageSizeSelector({
  pageSize,
  options = [10, 25, 50, 100],
  onPageSizeChange,
  label = 'Rows per page:',
  className,
  ...props
}: PageSizeSelectorProps) {
  return (
    <div
      className={cn('flex items-center gap-2 text-sm', className)}
      {...props}
    >
      <label className="text-[var(--text-muted)]">{label}</label>
      <select
        value={pageSize}
        onChange={(e) => onPageSizeChange(Number(e.target.value))}
        className={cn(
          'px-2 py-1 rounded-[var(--radius-input)]',
          'bg-[var(--surface-input)] border border-[var(--border-default)]',
          'text-[var(--text-primary)]',
          'focus:outline-none focus:ring-2 focus:ring-[var(--interactive-primary)] focus:border-transparent'
        )}
      >
        {options.map((option) => (
          <option key={option} value={option}>
            {option}
          </option>
        ))}
      </select>
    </div>
  );
}

/* ============================================
   PAGINATION HOOK
   ============================================ */

export interface UsePaginationOptions {
  /** Total number of items */
  totalItems: number;
  /** Initial page size */
  initialPageSize?: number;
  /** Initial page */
  initialPage?: number;
}

export interface UsePaginationReturn {
  currentPage: number;
  pageSize: number;
  totalPages: number;
  setPage: (page: number) => void;
  setPageSize: (size: number) => void;
  nextPage: () => void;
  prevPage: () => void;
  startIndex: number;
  endIndex: number;
}

export function usePagination({
  totalItems,
  initialPageSize = 10,
  initialPage = 1,
}: UsePaginationOptions): UsePaginationReturn {
  const [currentPage, setCurrentPage] = useState(initialPage);
  const [pageSize, setPageSizeState] = useState(initialPageSize);

  const totalPages = Math.ceil(totalItems / pageSize);

  const setPage = useCallback((page: number) => {
    setCurrentPage(Math.max(1, Math.min(page, totalPages)));
  }, [totalPages]);

  const setPageSize = useCallback((size: number) => {
    setPageSizeState(size);
    // Reset to first page when page size changes
    setCurrentPage(1);
  }, []);

  const nextPage = useCallback(() => {
    setPage(currentPage + 1);
  }, [currentPage, setPage]);

  const prevPage = useCallback(() => {
    setPage(currentPage - 1);
  }, [currentPage, setPage]);

  const startIndex = (currentPage - 1) * pageSize;
  const endIndex = Math.min(startIndex + pageSize - 1, totalItems - 1);

  return {
    currentPage,
    pageSize,
    totalPages,
    setPage,
    setPageSize,
    nextPage,
    prevPage,
    startIndex,
    endIndex,
  };
}
