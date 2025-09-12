import React, { useCallback, useMemo } from 'react';

import { Empty } from '@/components/ui/empty';
import { Skeleton } from '@/components/ui/skeleton';
import {
  Table,
  TableBody,
  TableCaption,
  TableCell,
  TableHead,
  TableHeader,
  TableRow
} from '@/components/ui/table';
import { cn } from '@/lib/utils';

export interface ColumnType<T> {
  title: React.ReactNode | (() => React.ReactNode);
  key: string;
  width?: string | number;
  className?: string;
  style?: React.CSSProperties;
  render?: (value: T, index: number) => React.ReactNode;
}

export interface CustomTableProps<T> {
  columns: ColumnType<T>[];
  dataSource: T[];
  rowKey: keyof T | ((record: T) => string);
  caption?: React.ReactNode;
  isLoading?: boolean;
  loadingRows?: number;
  loadingHeight?: number;
  emptyText?: React.ReactNode;
  className?: string;
  bodyClassName?: string;
  tableClassName?: string;
  maxHeight?: string;
  onRow?: (record: T, index: number) => React.HTMLAttributes<HTMLTableRowElement>;
}

export function CustomTable<T extends Record<string, unknown>>({
  columns,
  dataSource,
  rowKey,
  caption,
  isLoading = false,
  loadingRows = 5,
  loadingHeight = 30,
  emptyText = 'No data',
  className,
  bodyClassName,
  tableClassName,
  maxHeight = 'calc(100vh-200px)',
  onRow
}: CustomTableProps<T>) {
  const getRowKey = useCallback(
    (record: T): string => {
      if (typeof rowKey === 'function') {
        return rowKey(record);
      }
      return String(record[rowKey]);
    },
    [rowKey]
  );

  const renderCell = (record: T, column: ColumnType<T>, index: number) => {
    if (column.render) {
      return column.render(record, index);
    }

    const value = record[column.key as keyof T];
    return value !== null && value !== undefined ? String(value) : '';
  };

  const LoadingRows = useMemo(() => {
    return Array.from({ length: loadingRows }).map((_, index) => (
      <TableRow key={`loading-${index}`}>
        {columns.map((column) => (
          <TableCell
            key={`loading-cell-${column.key}-${index}`}
            className={column.className}
            style={{ width: column.width }}
          >
            <Skeleton
              className="w-full"
              style={{
                height: loadingHeight
              }}
            />
          </TableCell>
        ))}
      </TableRow>
    ));
  }, [columns, loadingRows, loadingHeight]);

  return (
    <div className={cn('bg-card rounded-[14px] p-[20px]', className)}>
      <Table className={cn(tableClassName)}>
        <TableHeader>
          <TableRow className="bg-muted !border-b-0">
            {columns.map((column, index) => (
              <TableHead
                key={column.key}
                className={cn(
                  'px-[20px] py-[15px] text-[12px] font-semibold',
                  index === 0 && 'rounded-l-[12px]',
                  index === columns.length - 1 && 'rounded-r-[12px]',
                  column.className
                )}
                style={{ width: column.width }}
              >
                {typeof column.title === 'function' ? column.title() : column.title}
              </TableHead>
            ))}
          </TableRow>
        </TableHeader>
      </Table>

      <div className={cn('custom-scrollbar overflow-y-auto', bodyClassName)} style={{ maxHeight }}>
        <Table className={cn(tableClassName)}>
          {caption && !!dataSource?.length && (
            <TableCaption className="m-0 px-[10px] pt-[20px] pb-0 text-[14px] font-normal">
              {caption}
            </TableCaption>
          )}
          <TableBody>
            {isLoading
              ? LoadingRows
              : dataSource?.length > 0
                ? dataSource.map((record, index) => {
                    const rowProps = onRow ? onRow(record, index) : {};
                    return (
                      <TableRow key={getRowKey(record)} {...rowProps}>
                        {columns.map((column) => (
                          <TableCell
                            key={`${getRowKey(record)}-${column.key}`}
                            className={cn('p-[20px] text-[16px] font-normal', column.className)}
                            style={{
                              width: column.width,
                              ...(column?.style || {})
                            }}
                          >
                            {renderCell(record, column, index)}
                          </TableCell>
                        ))}
                      </TableRow>
                    );
                  })
                : null}
          </TableBody>
        </Table>
      </div>

      {!isLoading && dataSource?.length === 0 && (
        <Empty
          label={emptyText}
          style={{
            height: loadingHeight * 4
          }}
        />
      )}
    </div>
  );
}
