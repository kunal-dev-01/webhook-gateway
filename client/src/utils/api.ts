export function unwrapResponse<T>(response: { data: { success: boolean; data: T } }): T {
  const data = response.data?.data;
  if (data === undefined || data === null) {
    throw new Error("Invalid API response structure");
  }
  return data;
}