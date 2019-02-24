export function assert(val: any, message: string) {
  if (!val) {
    throw new Error(`AssertionError: ${message}`)
  }
}