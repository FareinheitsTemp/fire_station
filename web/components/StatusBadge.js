export default function StatusBadge({ status }) {
  let mod = ''
  if (status === 'завершений' || status === 'в строю' || status === 'справно') mod = ' badge--ok'
  else if (status === 'в роботі' || status === 'ремонт') mod = ' badge--warn'
  else if (status === 'новий' || status === 'списано' || status === 'несправно') mod = ' badge--err'
  return <span className={`badge${mod}`}>{status}</span>
}
