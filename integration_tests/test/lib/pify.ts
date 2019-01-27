export function pify<T>(actor: (cb: (err: any, arg?: T) => void) => void): Promise<T> {
  return new Promise<T>(((resolve, reject) => actor((err, arg) => {
    if (err) {
      return reject(err);
    }

    resolve(arg);
  })));
}