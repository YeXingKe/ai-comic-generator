/** 相同 key 的并发请求复用同一 Promise，避免 StrictMode 等场景重复打接口 */
const inFlight = new Map<string, Promise<unknown>>()

export function dedupeRequest<T>(key: string, fn: () => Promise<T>): Promise<T> {
  const pending = inFlight.get(key)
  if (pending) {
    return pending as Promise<T>
  }

  const promise = fn().finally(() => {
    inFlight.delete(key)
  })
  inFlight.set(key, promise)
  return promise
}
