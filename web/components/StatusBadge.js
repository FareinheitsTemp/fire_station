export default function StatusBadge({ status }) {
  let mod = ''
  if (status === 'завершений') mod = ' badge--ok'
  else if (status === 'в роботі' || status === 'ремонт') mod = ' badge--warn'
  else if (status === 'новий') mod = ' badge--err'
  return <span className={`badge${mod}`}>{status}</span>
}
