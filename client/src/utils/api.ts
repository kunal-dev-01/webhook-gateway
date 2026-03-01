export function unwrapResponse<T>(response: { data: { success: boolean; data: T } }): T {
  const data = response.data?.data;
  return data;
}