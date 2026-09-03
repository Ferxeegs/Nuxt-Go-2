export type MediaUploadResult = {
  id: number
  file_name: string
  path: string
  url: string
  mime_type: string
  size: number
  related_module: string
  created_at: string
}

export function useMediaUpload() {
  const { api, accessToken, baseURL, refreshAccessToken, clearSessionAndRedirect } = useApi()

  async function uploadFile(file: File, module = 'general'): Promise<MediaUploadResult> {
    const form = new FormData()
    form.append('file', file)
    form.append('module', module)

    const doUpload = async (retried = false): Promise<MediaUploadResult> => {
      const headers: Record<string, string> = {}
      if (accessToken.value) {
        headers.Authorization = `Bearer ${accessToken.value}`
      }

      try {
        const res = await $fetch<{
          status: string
          message: string
          data?: MediaUploadResult
        }>(`${baseURL.value}/admin/media/upload`, {
          method: 'POST',
          body: form,
          headers,
        })
        if (!res.data?.url) {
          throw new Error(res.message || 'Upload failed')
        }
        return res.data
      } catch (err: unknown) {
        const status =
          (err as { statusCode?: number; status?: number })?.statusCode ??
          (err as { statusCode?: number; status?: number })?.status
        if (status === 401 && !retried) {
          const ok = await refreshAccessToken()
          if (ok) {
            return doUpload(true)
          }
          await clearSessionAndRedirect()
        }
        const e = err as { data?: { message?: string }; message?: string }
        throw new Error(e?.data?.message || e?.message || 'Upload failed')
      }
    }

    return doUpload()
  }

  return { uploadFile, api }
}
