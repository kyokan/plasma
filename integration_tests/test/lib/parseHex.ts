export function parseHex(hex: string) {
  return Buffer.from(hex.indexOf('0x') > -1 ? hex.slice(2) : hex, 'hex');
}

export function toHex(buf: Buffer) {
  return `0x${buf.toString('hex')}`
}