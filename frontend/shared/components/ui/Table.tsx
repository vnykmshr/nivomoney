/**
 * Table Component
 * Reusable data table with sorting support and responsive design
 */

import { useState, useCallback } from 'react';
import type { ReactNode, HTMLAttributes } from 'react';
import { cn } from '../../lib/utils';

/* ============================================
   TABLE ROOT
   ============================================ */

export interface TableProps extends HTMLAttributes<HTMLTableElement> {
  /** Make table horizontally scrollable on small screens */
  responsive?: boolean;
}

export function Table({ className, responsive = true, children, ...props }: TableProps) {
  const table = (
    <table
      className={cn(
        'w-full border-collapse text-sm',
        className
      )}
      {...props}
    >
      {children}
    </table>
  );

  if (responsive) {
    return (
      <div className="overflow-x-auto -mx-4 px-4 md:mx-0 md:px-0">
        {table}
      </div>
    );
  }

  return table;
}

/* ============================================
   TABLE HEAD
   ============================================ */

export interface TableHeadProps extends HTMLAttributes<HTMLTableSectionElement> {}

export function TableHead({ className, children, ...props }: TableHeadProps) {
  return (
    <thead
      className={cn(
        'bg-[var(--surface-secondary)]',
        className
      )}
      {...props}
    >
      {children}
    </thead>
  );
}

/* ============================================
   TABLE BODY
   ============================================ */

export interface TableBodyProps extends HTMLAttributes<HTMLTableSectionElement> {}

export function TableBody({ className, children, ...props }: TableBodyProps) {
  return (
    <tbody
      className={cn('divide-y divide-[var(--border-subtle)]', className)}
      {...props}
    >
      {children}
    </tbody>
  );
}

/* ============================================
   TABLE ROW
   ============================================ */

export interface TableRowProps extends HTMLAttributes<HTMLTableRowElement> {
  /** Highlight row on hover */
  hoverable?: boolean;
  /** Mark row as selected */
  selected?: boolean;
}

export function TableRow({ className, hoverable = true, selected, children, ...props }: TableRowProps) {
  return (
    <tr
      className={cn(
        'border-b border-[var(--border-subtle)] last:border-b-0',
        hoverable && 'hover:bg-[var(--surface-secondary)] transition-colors',
        selected && 'bg-[var(--surface-brand-subtle)]',
        className
      )}
      {...props}
    >
      {children}
    </tr>
  );
}

/* ============================================
   TABLE HEADER CELL
   ============================================ */

export type SortDirection = 'asc' | 'desc' | null;

export interface TableHeaderCellProps extends HTMLAttributes<HTMLTableCellElement> {
  /** Enable sorting for this column */
  sortable?: boolean;
  /** Current sort direction */
  sortDirection?: SortDirection;
  /** Callback when sort is clicked */
  onSort?: () => void;
  /** Alignment */
  align?: 'left' | 'center' | 'right';
}

export function TableHeaderCell({
  className,
  sortable,
  sortDirection,
  onSort,
  align = 'left',
  children,
  ...props
}: TableHeaderCellProps) {
  const alignClass = {
    left: 'text-left',
    center: 'text-center',
    right: 'text-right',
  }[align];

  const content = (
    <>
      <span>{children}</span>
      {sortable && (
        <span className="ml-1 inline-flex flex-col" aria-hidden="true">
          <svg
            className={cn(
              'w-3 h-3 -mb-1',
              sortDirection === 'asc' ? 'text-[var(--interactive-primary)]' : 'text-[var(--text-muted)]'
            )}
            fill="currentColor"
            viewBox="0 0 20 20"
          >
            <path d="M10 3l-7 7h14l-7-7z" />
          </svg>
          <svg
            className={cn(
              'w-3 h-3',
              sortDirection === 'desc' ? 'text-[var(--interactive-primary)]' : 'text-[var(--text-muted)]'
            )}
            fill="currentColor"
            viewBox="0 0 20 20"
          >
            <path d="M10 17l7-7H3l7 7z" />
          </svg>
        </span>
      )}
    </>
  );

  return (
    <th
      className={cn(
        'px-4 py-3 font-semibold text-[var(--text-secondary)]',
        alignClass,
        sortable && 'cursor-pointer select-none hover:text-[var(--text-primary)]',
        className
      )}
      onClick={sortable ? onSort : undefined}
      aria-sort={sortDirection === 'asc' ? 'ascending' : sortDirection === 'desc' ? 'descending' : undefined}
      {...props}
    >
      {sortable ? (
        <div className="inline-flex items-center gap-1">
          {content}
        </div>
      ) : (
        children
      )}
    </th>
  );
}

/* ============================================
   TABLE CELL
   ============================================ */

export interface TableCellProps extends HTMLAttributes<HTMLTableCellElement> {
  /** Alignment */
  align?: 'left' | 'center' | 'right';
  /** Truncate text with ellipsis */
  truncate?: boolean;
}

export function TableCell({
  className,
  align = 'left',
  truncate,
  children,
  ...props
}: TableCellProps) {
  const alignClass = {
    left: 'text-left',
    center: 'text-center',
    right: 'text-right',
  }[align];

  return (
    <td
      className={cn(
        'px-4 py-3 text-[var(--text-primary)]',
        alignClass,
        truncate && 'truncate max-w-[200px]',
        className
      )}
      {...props}
    >
      {children}
    </td>
  );
}

/* ============================================
   TABLE FOOTER
   ============================================ */

export interface TableFooterProps extends HTMLAttributes<HTMLTableSectionElement> {}

export function TableFooter({ className, children, ...props }: TableFooterProps) {
  return (
    <tfoot
      className={cn(
        'bg-[var(--surface-secondary)] border-t border-[var(--border-default)]',
        className
      )}
      {...props}
    >
      {children}
    </tfoot>
  );
}

/* ============================================
   EMPTY STATE
   ============================================ */

export interface TableEmptyProps {
  /** Number of columns to span */
  colSpan: number;
  /** Icon to display */
  icon?: ReactNode;
  /** Title text */
  title?: string;
  /** Description text */
  description?: string;
  /** Action button/link */
  action?: ReactNode;
}

export function TableEmpty({
  colSpan,
  icon,
  title = 'No data',
  description,
  action,
}: TableEmptyProps) {
  return (
    <tr>
      <td colSpan={colSpan} className="px-4 py-12 text-center">
        {icon && (
          <div className="w-12 h-12 mx-auto mb-4 rounded-full bg-[var(--surface-secondary)] flex items-center justify-center text-[var(--text-muted)]">
            {icon}
          </div>
        )}
        <h3 className="text-lg font-medium text-[var(--text-primary)] mb-1">{title}</h3>
        {description && (
          <p className="text-sm text-[var(--text-muted)] mb-4">{description}</p>
        )}
        {action}
      </td>
    </tr>
  );
}

/* ============================================
   SORTABLE TABLE HOOK
   ============================================ */

export interface UseSortableTableOptions<T> {
  data: T[];
  defaultSortKey?: keyof T;
  defaultSortDirection?: SortDirection;
}

export interface UseSortableTableReturn<T> {
  sortedData: T[];
  sortKey: keyof T | null;
  sortDirection: SortDirection;
  requestSort: (key: keyof T) => void;
  getSortDirection: (key: keyof T) => SortDirection;
}

export function useSortableTable<T extends Record<string, unknown>>({
  data,
  defaultSortKey,
  defaultSortDirection = 'asc',
}: UseSortableTableOptions<T>): UseSortableTableReturn<T> {
  const [sortKey, setSortKey] = useState<keyof T | null>(defaultSortKey ?? null);
  const [sortDirection, setSortDirection] = useState<SortDirection>(
    defaultSortKey ? defaultSortDirection : null
  );

  const requestSort = useCallback((key: keyof T) => {
    if (sortKey === key) {
      // Cycle through: asc -> desc -> null
      if (sortDirection === 'asc') {
        setSortDirection('desc');
      } else if (sortDirection === 'desc') {
        setSortDirection(null);
        setSortKey(null);
      } else {
        setSortDirection('asc');
      }
    } else {
      setSortKey(key);
      setSortDirection('asc');
    }
  }, [sortKey, sortDirection]);

  const getSortDirection = useCallback((key: keyof T): SortDirection => {
    return sortKey === key ? sortDirection : null;
  }, [sortKey, sortDirection]);

  const sortedData = [...data].sort((a, b) => {
    if (!sortKey || !sortDirection) return 0;

    const aVal = a[sortKey];
    const bVal = b[sortKey];

    if (aVal === bVal) return 0;
    if (aVal === null || aVal === undefined) return 1;
    if (bVal === null || bVal === undefined) return -1;

    let comparison = 0;
    if (typeof aVal === 'string' && typeof bVal === 'string') {
      comparison = aVal.localeCompare(bVal);
    } else if (typeof aVal === 'number' && typeof bVal === 'number') {
      comparison = aVal - bVal;
    } else if (aVal instanceof Date && bVal instanceof Date) {
      comparison = aVal.getTime() - bVal.getTime();
    } else {
      comparison = String(aVal).localeCompare(String(bVal));
    }

    return sortDirection === 'asc' ? comparison : -comparison;
  });

  return {
    sortedData,
    sortKey,
    sortDirection,
    requestSort,
    getSortDirection,
  };
}
