import { useRouter } from 'next/router'
import { useEffect, useState } from 'react'
import { get } from '../../lib/api'
import DataTable from '../../components/DataTable'

export default function TableView() {
  const router = useRouter()
  const { name } = router.query
  const [data, setData] = useState(null)
  const [error, setError] = useState('')

  useEffect(() => {
    if (!name) return
    get(`/api/tables/${name}`)
      .then(setData)
      .catch((e) => setError(e.message))
  }, [name])

  return (
    <div className="page">
      <h1 className="page__title">Таблиця: {name}</h1>
      {error && <div className="alert">{error}</div>}
      {data && <DataTable columns={data.columns} rows={data.rows} />}
    </div>
  )
}
