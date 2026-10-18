import type { ReactNode } from "react";

export type Column<T> = {
  header: string;
  cell: (row: T) => ReactNode;
};

export type DataTableProps<T> = {
  columns: Column<T>[];
  rows: T[];
  rowKey: (row: T) => string;
  empty?: string;
};

/** A plain table; shows `empty` instead of an empty body. */
export function DataTable<T>({ columns, rows, rowKey, empty = "Nothing here yet." }: DataTableProps<T>) {
  if (rows.length === 0) {
    return <p className="sk-empty">{empty}</p>;
  }
  return (
    <table className="sk-table">
      <thead>
        <tr>
          {columns.map((c) => (
            <th key={c.header}>{c.header}</th>
          ))}
        </tr>
      </thead>
      <tbody>
        {rows.map((row) => (
          <tr key={rowKey(row)}>
            {columns.map((c) => (
              <td key={c.header}>{c.cell(row)}</td>
            ))}
          </tr>
        ))}
      </tbody>
    </table>
  );
}
