// PHP's number_format($n, $decimals): rounds half away from zero to
// `decimals` places, "," thousands separator, "." decimal point.
export function numberFormat(value, decimals = 0) {
  const n = Number(value) || 0
  const factor = 10 ** decimals
  const rounded = Math.sign(n) * Math.round(Math.abs(n) * factor + Number.EPSILON) / factor
  const [int, frac] = Math.abs(rounded).toFixed(decimals).split('.')
  const sign = rounded < 0 ? '-' : ''
  return sign + int.replace(/\B(?=(\d{3})+(?!\d))/g, ',') + (frac ? '.' + frac : '')
}

// The Go backend's models.Null* types serialise flat (a plain value or
// null); these also accept database/sql's {String|Int64|Time, Valid} shape.
function unwrapNull(v, key) {
  if (v == null) return null
  if (typeof v === 'object' && 'Valid' in v) return v.Valid ? v[key] : null
  return v
}
export const nullStr = (v) => unwrapNull(v, 'String') ?? ''
export const nullInt = (v) => unwrapNull(v, 'Int64')
export const nullTime = (v) => unwrapNull(v, 'Time')

const pad = (n) => String(n).padStart(2, '0')

// Carbon's default string cast ("Y-m-d H:i:s"). Timestamps from the API
// carry the database's wall-clock time, which is what PHP printed, so that
// is kept as-is rather than converted to the browser's time zone.
// `length` 16 gives Carbon's format('Y-m-d H:i').
export function phpDateTime(value, length = 19) {
  if (!value) return ''
  const m = String(value).match(/^(\d{4}-\d{2}-\d{2})[T ](\d{2}:\d{2}:\d{2})/)
  if (m) return `${m[1]} ${m[2]}`.slice(0, length)
  const d = new Date(value)
  if (Number.isNaN(d.getTime())) return ''
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

// Carbon's diffForHumans() relative to now ("3 hours ago", "1 day ago").
export function diffForHumans(value) {
  if (!value) return ''
  const then = new Date(value).getTime()
  if (Number.isNaN(then)) return ''
  const seconds = Math.round((Date.now() - then) / 1000)
  const abs = Math.abs(seconds)
  const units = [
    ['year', 31536000],
    ['month', 2629746],
    ['week', 604800],
    ['day', 86400],
    ['hour', 3600],
    ['minute', 60],
    ['second', 1],
  ]
  for (const [name, secs] of units) {
    const n = Math.floor(abs / secs)
    if (n >= 1 || name === 'second') {
      const count = Math.max(n, 1)
      const text = `${count} ${name}${count === 1 ? '' : 's'}`
      return seconds >= 0 ? `${text} ago` : `${text} from now`
    }
  }
  return ''
}
