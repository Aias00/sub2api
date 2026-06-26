export function applyImageGeneratorDraft(
  draft: { prompt?: string; title?: string } | null | undefined,
  maxPromptLength: number,
) {
  if (!draft) {
    return { prompt: '', title: '' }
  }

  const prompt = draft.prompt?.trim() ? draft.prompt.trim().slice(0, maxPromptLength) : ''
  const title = draft.title?.trim() ? draft.title.trim() : ''
  return { prompt, title }
}

export function resolveImageGeneratorCatalogPath(catalogPath: string | null | undefined) {
  return catalogPath || ''
}
