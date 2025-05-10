export interface APIResponse<T = any> {
  code: number;
  message: string;
  data?: T;
  error?: any;
}

export function unwrapAPIResponse<T>(response: APIResponse<T>): T {
  if (response.error) {
    throw new Error(response.error);
  }
  if (response.data === undefined) {
    throw new Error('No data available in the response');
  }
  return response.data;
}
