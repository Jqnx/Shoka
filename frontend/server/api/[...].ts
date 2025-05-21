import { joinURL } from 'ufo'

export default defineEventHandler(async (event) => {
  const apiUrl = useRuntimeConfig().baseApi

  const path = event.path.replace(/^\/api\//, '')

  const target = joinURL(apiUrl, path)

  return proxyRequest(event, target)
})