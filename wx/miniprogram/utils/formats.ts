export namespace formats {
  export function formatDuration(sec: number) {
    const padStr = (n: number) => {
      return n.toFixed(0).padStart(2, '0')
    }
    const h = Math.floor(sec / 3600)
    sec -= h * 3600
    const m = Math.floor(sec / 60)
    sec -= m * 60
    const s = Math.floor(sec)

    return {
      hh: padStr(h),
      mm: padStr(m),
      ss: padStr(s),
    }
  }

  export function formatFee(cents: number | string) {
    return `${(Number(cents) / 100).toFixed(2)}元`
  }
}
