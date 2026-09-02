export default function DataTable({ columns = [], rows = [], actions }) {
  if (!rows.length) {
    return <p className="muted">(порожньо)</p>
  }
  return (
    <div className="table-wrap">
      <table className="table">
        <thead>
          <tr>
            {columns.map((c) => (
              <th key={c}>{c}</th>
            ))}
            {actions && <th>Дії</th>}
          </tr>
        </thead>
        <tbody>
          {rows.map((row, i) => (
            <tr key={i}>
              {row.map((cell, j) => (
                <td key={j}>{cell}</td>
              ))}
              {actions && <td className="table__actions">{actions(row, i)}</td>}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
