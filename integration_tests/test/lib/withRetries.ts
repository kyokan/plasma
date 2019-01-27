import {wait} from './wait';

export async function withRetryCondition<T> (cb: () => Promise<T>, cond: (val: T) => boolean, retryCount: number = 10, retryTime: number = 1000) {
  return withRetries(async () => {
    const res = await cb();
    if (cond(res)) {
      return res;
    }

    throw new Error('condition not met');
  }, retryCount, retryTime);
}

export async function withRetries<T> (cb: () => Promise<T>, retryCount: number = 10, retryTime: number = 1000) {
  for (let i = 0; i < retryCount; i++) {
    try {
      return await cb();
    } catch (e) {
      await wait(retryTime);
    }
  }

  throw new Error('Timed out.');
}